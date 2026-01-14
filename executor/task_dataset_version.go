package executor

import (
	"context"
	"strconv"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/mapper"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// DatasetVersionTaskExecutor executes migration tasks for dataset editions.
type DatasetVersionTaskExecutor struct {
	jobService       application.JobService
	clientList       *clients.ClientList
	serviceAuthToken string
}

// NewDatasetVersionTaskExecutor creates a new DatasetVersionTaskExecutor
func NewDatasetVersionTaskExecutor(jobService application.JobService, clientList *clients.ClientList, serviceAuthToken string) *DatasetVersionTaskExecutor {
	return &DatasetVersionTaskExecutor{
		jobService:       jobService,
		clientList:       clientList,
		serviceAuthToken: serviceAuthToken,
	}
}

// Migrate handles the migration operations for a dataset edition task.
func (e *DatasetVersionTaskExecutor) Migrate(ctx context.Context, task *domain.Task) error {
	logData := log.Data{"task_id": task.ID, "job_number": task.JobNumber, "source_id": task.Source.ID}

	log.Info(ctx, "starting migration for dataset version task", logData)

	sourceData, err := e.clientList.Zebedee.GetDataset(ctx, e.serviceAuthToken, zebedee.EmptyCollectionId, zebedee.EnglishLangCode, task.Source.ID)
	if err != nil {
		log.Error(ctx, "failed to get source version data from zebedee", err, logData)
		return err
	}

	// Usage notes only appear at the series level so we need that too.
	seriesData, err := e.clientList.Zebedee.GetDatasetLandingPage(ctx, e.serviceAuthToken, zebedee.EmptyCollectionId, zebedee.EnglishLangCode, task.Target.DatasetID)
	if err != nil {
		logData["dataset_id"] = task.Target.DatasetID
		log.Error(ctx, "failed to get dataset series data from zebedee", err, logData)
		return err
	}

	// Correction notes only appear at the edition level so we need that too.
	editionData, err := e.clientList.Zebedee.GetDataset(ctx, e.serviceAuthToken, zebedee.EmptyCollectionId, zebedee.EnglishLangCode, task.Target.EditionID)
	if err != nil {
		logData["edition_id"] = task.Target.EditionID
		log.Error(ctx, "failed to get edition data from zebedee", err, logData)
		return err
	}

	datasetVersion, err := mapper.MapDatasetVersionToDatasetAPI(task.Target.EditionID, sourceData, seriesData, editionData)
	if err != nil {
		log.Error(ctx, "failed to map dataset version to dataset API model", err, logData)
		return err
	}

	versionID := datasetVersion.Version
	versionIDStr := strconv.Itoa(versionID)

	logData["version_id"] = versionID
	if task.Target != nil {
		task.Target.ID = versionIDStr
	} else {
		task.Target = &domain.TaskMetadata{
			ID: versionIDStr,
		}
	}

	err = e.jobService.UpdateTask(ctx, task)
	if err != nil {
		log.Error(ctx, "failed to update version migration task with target id", err, logData)
		return err
	}

	headers := sdk.Headers{
		AccessToken: e.serviceAuthToken,
	}

	log.Info(ctx, "creating dataset version in dataset API", logData)

	_, err = e.clientList.DatasetAPI.PostVersion(ctx, headers, task.Target.DatasetID, task.Target.EditionID, task.Target.ID, *datasetVersion)
	if err != nil {
		log.Error(ctx, "failed to create target dataset version in dataset API", err, logData)
		return err
	}

	// Create download tasks for each download in the source data
	for _, download := range sourceData.Downloads {
		downloadTask := domain.NewTask(task.JobNumber)

		downloadTask.Type = domain.TaskTypeDatasetDownload
		downloadTask.Source = &domain.TaskMetadata{
			ID: task.Source.ID + "/" + download.File,
		}
		downloadTask.Target = &domain.TaskMetadata{
			DatasetID: task.Target.DatasetID,
			EditionID: task.Target.EditionID,
			VersionID: versionIDStr,
		}

		_, err := e.jobService.CreateTask(ctx, task.JobNumber, &downloadTask)
		if err != nil {
			logData["version_uri"] = downloadTask.Source.ID
			log.Error(ctx, "failed to create migration task for download task for version", err, logData)
			return err
		}
	}

	// Mark task as complete
	err = e.jobService.UpdateTaskState(ctx, task.ID, domain.StateInReview)
	if err != nil {
		log.Error(ctx, "failed to update migration task", err, logData)
		return err
	}
	log.Info(ctx, "completed migration for dataset version task", logData)
	return nil
}

// Publish handles the publish operations for a dataset edition task.
func (e *DatasetVersionTaskExecutor) Publish(ctx context.Context, task *domain.Task) error {
	// Implementation of publish for dataset edition
	return nil
}

// PostPublish handles the post-publish operations for a dataset edition task.
func (e *DatasetVersionTaskExecutor) PostPublish(ctx context.Context, task *domain.Task) error {
	// Implementation of post-publish for dataset edition
	return nil
}

// Revert handles the revert operations for a dataset edition task.
func (e *DatasetVersionTaskExecutor) Revert(ctx context.Context, task *domain.Task) error {
	// Implementation of revert for dataset edition
	return nil
}
