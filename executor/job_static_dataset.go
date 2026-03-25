package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetSDK "github.com/ONSdigital/dp-dataset-api/sdk"
	filesSDK "github.com/ONSdigital/dp-files-api/sdk"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	"github.com/ONSdigital/log.go/v2/log"
)

var deleteDatasetFromAPI = func(ctx context.Context, datasetAPIURL, datasetID, serviceAuthToken string) error {
	if datasetAPIURL == "" || datasetID == "" {
		return fmt.Errorf("dataset api url and dataset id are required")
	}

	endpoint := strings.TrimRight(datasetAPIURL, "/") + "/datasets/" + url.PathEscape(datasetID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, http.NoBody)
	if err != nil {
		return err
	}

	dprequest.AddServiceTokenHeader(req, serviceAuthToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete dataset %q: status %d body %q", datasetID, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

// StaticDatasetJobExecutor executes migration jobs for static datasets.
type StaticDatasetJobExecutor struct {
	clientList       *clients.ClientList
	jobService       application.JobService
	serviceAuthToken string
}

// NewStaticDatasetJobExecutor creates a new StaticDatasetJobExecutor
func NewStaticDatasetJobExecutor(jobService application.JobService, clientList *clients.ClientList, serviceAuthToken string) *StaticDatasetJobExecutor {
	return &StaticDatasetJobExecutor{
		jobService:       jobService,
		clientList:       clientList,
		serviceAuthToken: serviceAuthToken,
	}
}

// Migrate handles the migration operations for a static dataset job.
func (e *StaticDatasetJobExecutor) Migrate(ctx context.Context, job *domain.Job) error {
	logData := log.Data{"job_number": job.JobNumber}
	log.Info(ctx, "starting migration for job", logData)

	datasetSeriesTask := domain.NewTask(job.JobNumber)

	datasetSeriesTask.Type = domain.TaskTypeDatasetSeries
	datasetSeriesTask.Source = &domain.TaskMetadata{
		ID: job.Config.SourceID,
	}

	datasetSeriesTask.Target = &domain.TaskMetadata{
		ID: job.Config.TargetID,
	}

	collection := domain.NewMigrationCollection(job.JobNumber)

	logData["collection_name"] = collection.Name
	log.Info(ctx, "creating collection for migration job", logData)

	createdCollection, err := e.clientList.Zebedee.CreateCollection(ctx, e.serviceAuthToken, collection)
	if err != nil {
		log.Error(ctx, "failed to create collection for migration job", err, logData)
		return err
	}

	logData["collection_id"] = createdCollection.ID
	log.Info(ctx, "updating job with collection id", logData)

	err = e.jobService.UpdateJobCollectionID(ctx, job.JobNumber, createdCollection.ID)
	if err != nil {
		log.Error(ctx, "failed to update job collection ID", err, logData)
		return err
	}

	_, err = e.jobService.CreateTask(ctx, job.JobNumber, &datasetSeriesTask)
	if err != nil {
		logData["task_source_id"] = datasetSeriesTask.Source.ID
		log.Error(ctx, "failed to create migration task", err, logData)
		return err
	}

	return nil
}

// Publish handles the publish operations for a static dataset job.
func (e *StaticDatasetJobExecutor) Publish(ctx context.Context, job *domain.Job) error {
	// Implementation of publish for static dataset
	return nil
}

// PostPublish handles the post-publish operations for a static dataset job.
func (e *StaticDatasetJobExecutor) PostPublish(ctx context.Context, job *domain.Job) error {
	// Implementation of post-publish for static dataset
	return nil
}

// Revert handles the revert operations for a static dataset job.
func (e *StaticDatasetJobExecutor) Revert(ctx context.Context, job *domain.Job) error {
	if job == nil || job.Config == nil {
		return fmt.Errorf("job and job config are required for revert")
	}

	logData := log.Data{
		"job_number": job.JobNumber,
		"target_id":  job.Config.TargetID,
	}

	log.Info(ctx, "starting revert for static dataset job", logData)

	if err := e.revertDownloadFiles(ctx, job.JobNumber); err != nil {
		log.Error(ctx, "failed to revert download files", err, logData)
		return err
	}

	if err := deleteDatasetFromAPI(ctx, e.clientList.DatasetAPI.URL(), job.Config.TargetID, e.serviceAuthToken); err != nil {
		log.Error(ctx, "failed to delete dataset during job revert", err, logData)
		return err
	}

	if job.Config.CollectionID != "" {
		log.Warn(ctx, "collection deletion not added yet", log.Data{
			"job_number":    job.JobNumber,
			"collection_id": job.Config.CollectionID,
		})
	}

	log.Info(ctx, "completed revert for static dataset job", logData)
	return nil
}

func (e *StaticDatasetJobExecutor) revertDownloadFiles(ctx context.Context, jobNumber int) error {
	total, err := e.jobService.CountTasksByJobNumber(ctx, jobNumber)
	if err != nil {
		return err
	}

	if total == 0 {
		return nil
	}

	tasks, _, err := e.jobService.GetJobTasks(ctx, nil, jobNumber, total, 0)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.Type != domain.TaskTypeDatasetDownload {
			continue
		}

		if err := e.revertDownloadTask(ctx, task); err != nil {
			return err
		}
	}

	return nil
}

func (e *StaticDatasetJobExecutor) revertDownloadTask(ctx context.Context, task *domain.Task) error {
	if task == nil || task.Target == nil {
		return nil
	}

	if task.Target.DatasetID == "" || task.Target.EditionID == "" || task.Target.VersionID == "" {
		return nil
	}

	headers := datasetSDK.Headers{AccessToken: e.serviceAuthToken}
	version, responseHeaders, err := e.clientList.DatasetAPI.GetVersionWithHeaders(
		ctx,
		headers,
		task.Target.DatasetID,
		task.Target.EditionID,
		task.Target.VersionID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}

	fileName := filepath.Base(task.Source.ID)
	filePath, updatedDistributions, found := removeDistributionByTitle(version.Distributions, fileName)
	if !found {
		return nil
	}

	version.Distributions = &updatedDistributions
	headers.IfMatch = responseHeaders.ETag
	if _, err := e.clientList.DatasetAPI.PutVersion(
		ctx,
		headers,
		task.Target.DatasetID,
		task.Target.EditionID,
		task.Target.VersionID,
		version,
	); err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}

	if filePath != "" && e.clientList.FilesAPI != nil {
		deleteHeaders := filesSDK.Headers{Authorization: e.serviceAuthToken}
		if err := e.clientList.FilesAPI.DeleteFile(ctx, filePath, deleteHeaders); err != nil && !strings.Contains(err.Error(), "404") {
			log.Warn(ctx, "failed to delete upload file during revert", log.Data{"task_id": task.ID, "file_path": filePath, "error": err.Error()})
		}
	}

	return nil
}

func removeDistributionByTitle(distributions *[]datasetModels.Distribution, title string) (string, []datasetModels.Distribution, bool) {
	if distributions == nil {
		return "", nil, false
	}

	updated := make([]datasetModels.Distribution, 0, len(*distributions))
	deletedPath := ""
	found := false

	for _, distribution := range *distributions {
		if !found && distribution.Title == title {
			deletedPath = distribution.DownloadURL
			found = true
			continue
		}
		updated = append(updated, distribution)
	}

	return deletedPath, updated, found
}
