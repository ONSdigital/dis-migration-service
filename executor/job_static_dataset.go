package executor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	"github.com/ONSdigital/log.go/v2/log"
)

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
	logData := log.Data{"job_number": job.JobNumber}
	log.Info(ctx, "starting publishing for job", logData)

	totalTasks, err := e.jobService.CountTasksByJobNumber(ctx, job.JobNumber)
	if err != nil {
		log.Error(ctx, "failed to count tasks", err, logData)
		return err
	}

	tasks, _, err := e.jobService.GetJobTasks(ctx, []domain.State{domain.StateInReview}, job.JobNumber, totalTasks, 0)
	if err != nil {
		log.Error(ctx, "failed to get job tasks", err, logData)
		return err
	}
	log.Info(ctx, "successfully got all job tasks", logData)

	for _, task := range tasks {
		err := e.jobService.UpdateTaskState(ctx, task.ID, domain.StateApproved)
		if err != nil {
			log.Error(ctx, "failed to update task state", err, logData)
			return err
		}
	}
	log.Info(ctx, "successfully updated all job tasks state to approved", logData)

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
	revertErrors := make([]error, 0, 2)

	if err := e.deleteDatasetFromAPI(ctx, e.clientList.DatasetAPI.URL(), job.Config.TargetID, e.serviceAuthToken); err != nil {
		log.Error(ctx, "failed to delete dataset during job revert", err, logData)
		revertErrors = append(revertErrors, err)
	}

	if job.Config.CollectionID != "" {
		logData["collection_id"] = job.Config.CollectionID
		if err := e.revertZebedeeCollection(ctx, job.JobNumber, job.Config.CollectionID); err != nil {
			log.Error(ctx, "failed to delete zebedee collection during job revert", err, logData)
			revertErrors = append(revertErrors, err)
		}
	}

	if len(revertErrors) > 0 {
		return fmt.Errorf("failed to fully revert static dataset job: %w", errors.Join(revertErrors...))
	}

	log.Info(ctx, "completed revert for static dataset job", logData)
	return nil
}

func (e *StaticDatasetJobExecutor) deleteDatasetFromAPI(ctx context.Context, datasetAPIURL, datasetID, serviceAuthToken string) error {
	if datasetAPIURL == "" || datasetID == "" {
		return fmt.Errorf("dataset api url and dataset id are required")
	}

	endpoint := strings.TrimRight(datasetAPIURL, "/") + "/datasets/" + url.PathEscape(datasetID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, http.NoBody)
	if err != nil {
		return err
	}

	dprequest.AddServiceTokenHeader(req, serviceAuthToken)

	//nolint:gosec // G704 false positive: destination is derived from trusted Dataset API config.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("failed to delete dataset %q: status %d body %q", datasetID, resp.StatusCode, strings.TrimSpace(string(body)))
}

// revertZebedeeCollection removes collection content first
// then the collection.
func (e *StaticDatasetJobExecutor) revertZebedeeCollection(ctx context.Context, jobNumber int, collectionID string) error {
	if collectionID == "" || e.clientList.Zebedee == nil {
		return nil
	}

	contentPaths, err := e.getZebedeeContentPathsForRevert(ctx, jobNumber)
	if err != nil {
		return err
	}

	for _, contentPath := range contentPaths {
		deleteLogData := log.Data{
			"job_number":    jobNumber,
			"collection_id": collectionID,
			"content_path":  contentPath,
		}

		if err := e.clientList.Zebedee.DeleteCollectionContent(ctx, e.serviceAuthToken, collectionID, contentPath); err != nil {
			if !strings.Contains(err.Error(), "404") {
				log.Error(ctx, "failed to delete zebedee collection content", err, deleteLogData)
				return err
			}
			log.Info(ctx, "zebedee collection content not found or already deleted", deleteLogData)
			continue
		}

		log.Info(ctx, "successfully deleted zebedee collection content", deleteLogData)
	}

	if err := e.clientList.Zebedee.DeleteCollection(ctx, e.serviceAuthToken, collectionID); err != nil {
		if !strings.Contains(err.Error(), "404") {
			log.Error(ctx, "failed to delete zebedee collection", err)
			return err
		}
		log.Info(ctx, "zebedee collection not found or already deleted")
		return nil
	}

	log.Info(ctx, "successfully deleted zebedee collection")
	return nil
}

func (e *StaticDatasetJobExecutor) getZebedeeContentPathsForRevert(ctx context.Context, jobNumber int) ([]string, error) {
	total, err := e.jobService.CountTasksByJobNumber(ctx, jobNumber)
	if err != nil {
		return nil, err
	}

	if total == 0 {
		return nil, nil
	}

	tasks, _, err := e.jobService.GetJobTasks(ctx, nil, jobNumber, total, 0)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	paths := make([]string, 0)

	for _, task := range tasks {
		if task == nil || task.Source == nil || task.Source.ID == "" {
			continue
		}

		if task.Type != domain.TaskTypeDatasetSeries && task.Type != domain.TaskTypeDatasetVersion {
			continue
		}

		if _, ok := seen[task.Source.ID]; ok {
			continue
		}

		seen[task.Source.ID] = struct{}{}
		paths = append(paths, task.Source.ID)
	}

	return paths, nil
}
