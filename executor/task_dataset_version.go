package executor

import (
	"context"
	"strconv"
	"time"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/cache"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/mapper"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// These could be abstracted to config later if desired.
const (
	versionPublishPollMaxRetries = 10
	versionPublishPollDelay      = 500 * time.Millisecond
)

// DatasetVersionTaskExecutor executes migration tasks for dataset versions.
type DatasetVersionTaskExecutor struct {
	jobService       application.JobService
	clientList       *clients.ClientList
	serviceAuthToken string
	topicCache       *cache.TopicCache
}

// NewDatasetVersionTaskExecutor creates a new DatasetVersionTaskExecutor
func NewDatasetVersionTaskExecutor(jobService application.JobService, clientList *clients.ClientList, serviceAuthToken string, topicCache *cache.TopicCache) *DatasetVersionTaskExecutor {
	return &DatasetVersionTaskExecutor{
		jobService:       jobService,
		clientList:       clientList,
		serviceAuthToken: serviceAuthToken,
		topicCache:       topicCache,
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
	seriesData, err := e.clientList.Zebedee.GetDatasetLandingPage(ctx, e.serviceAuthToken, zebedee.EmptyCollectionId, zebedee.EnglishLangCode, task.Source.DatasetID)
	if err != nil {
		logData["dataset_id"] = task.Target.DatasetID
		log.Error(ctx, "failed to get dataset series data from zebedee", err, logData)
		return err
	}

	// Correction notes only appear at the edition level so we need that too.
	editionData, err := e.clientList.Zebedee.GetDataset(ctx, e.serviceAuthToken, zebedee.EmptyCollectionId, zebedee.EnglishLangCode, task.Source.EditionID)
	if err != nil {
		logData["edition_id"] = task.Target.EditionID
		log.Error(ctx, "failed to get edition data from zebedee", err, logData)
		return err
	}

	datasetVersion, err := mapper.MapDatasetVersionToDatasetAPI(task.Target.EditionID, task.Target.DatasetID, sourceData, seriesData, editionData)
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

	job, err := e.jobService.GetJob(ctx, task.JobNumber)
	if err != nil {
		log.Error(ctx, "failed to get job for dataset version task", err, logData)
		return err
	}

	logData["collection_id"] = job.Config.CollectionID

	datasetTopicSlug := cache.ExtractSingleTopicSlugFromURI(ctx, sourceData.URI, e.topicCache)
	datasetVersionLink := mapper.CreateDatasetVersionLink(datasetTopicSlug, datasetVersion)

	sourceData.Description.MigrationLink = datasetVersionLink
	err = e.clientList.Zebedee.SaveContentToCollection(
		ctx,
		e.serviceAuthToken,
		job.Config.CollectionID,
		task.Source.ID,
		sourceData,
	)
	if err != nil {
		log.Error(ctx, "failed to save updated dataset version page with migration link to zebedee collection", err, logData)
		return err
	}

	err = e.clientList.Zebedee.CompleteCollectionContent(
		ctx,
		e.serviceAuthToken,
		job.Config.CollectionID,
		zebedee.EnglishLangCode,
		task.Source.ID,
	)
	if err != nil {
		log.Error(ctx, "failed to complete dataset version content in zebedee collection", err, logData)
		return err
	}

	err = e.clientList.Zebedee.ApproveCollectionContent(
		ctx,
		e.serviceAuthToken,
		job.Config.CollectionID,
		zebedee.EnglishLangCode,
		task.Source.ID,
	)
	if err != nil {
		log.Error(ctx, "failed to approve dataset version content in zebedee collection", err, logData)
		return err
	}

	headers := sdk.Headers{
		AccessToken: e.serviceAuthToken,
	}

	log.Info(ctx, "creating dataset version in dataset API", logData)

	// Populating latest_version
	isLatest := false
	if len(seriesData.Datasets) > 0 {
		isLatest = editionData.URI == seriesData.Datasets[0].URI && datasetVersion.Version == len(editionData.Versions)+1
	}

	_, err = e.clientList.DatasetAPI.PostVersion(ctx, headers, task.Target.DatasetID, task.Target.EditionID, task.Target.ID, *datasetVersion, isLatest)
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

// Publish handles the publish operations for a dataset version task.
func (e *DatasetVersionTaskExecutor) Publish(ctx context.Context, task *domain.Task) error {
	logData := log.Data{"task_id": task.ID, "job_number": task.JobNumber}

	log.Info(ctx, "starting publish for dataset version task", logData)

	headers := sdk.Headers{
		AccessToken: e.serviceAuthToken,
	}

	err := e.clientList.DatasetAPI.PutVersionState(ctx, headers, task.Target.DatasetID, task.Target.EditionID, task.Target.ID, datasetModels.ApprovedState)
	if err != nil {
		log.Error(ctx, "failed to update dataset version state to approved", err, logData)
		return err
	}

	err = e.clientList.DatasetAPI.PutVersionState(ctx, headers, task.Target.DatasetID, task.Target.EditionID, task.Target.ID, datasetModels.PublishedState)
	if err != nil {
		log.Error(ctx, "failed to update dataset version state to published", err, logData)
		return err
	}

	for attempt := 1; attempt <= versionPublishPollMaxRetries; attempt++ {
		if attempt > 1 {
			log.Info(ctx, "version not yet published, retrying", log.Data{
				"attempt": attempt,
				"delay":   versionPublishPollDelay.String(),
			}, logData)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(versionPublishPollDelay):
			}
		}
		version, err := e.clientList.DatasetAPI.GetVersion(ctx, headers, task.Target.DatasetID, task.Target.EditionID, task.Target.ID)
		if err != nil {
			log.Error(ctx, "failed to get dataset version", err, logData)
			return err
		}

		if version.State == datasetModels.PublishedState {
			log.Info(ctx, "dataset version is now published", logData)
			break
		}
	}

	// Mark task as complete
	err = e.jobService.UpdateTaskState(ctx, task.ID, domain.StatePublished)
	if err != nil {
		log.Error(ctx, "failed to update migration task", err, logData)
		return err
	}
	log.Info(ctx, "completed migration for dataset version task", logData)
	return nil
}

// PostPublish handles the post-publish operations for a dataset version task.
func (e *DatasetVersionTaskExecutor) PostPublish(ctx context.Context, task *domain.Task) error {
	// Implementation of post-publish for dataset version
	logData := log.Data{"task_id": task.ID, "job_number": task.JobNumber}
	log.Info(ctx, "updating task state to completed", logData)

	err := e.jobService.UpdateTaskState(ctx, task.ID, domain.StateCompleted)
	if err != nil {
		log.Error(ctx, "failed to update task state to completed", err, logData)
		return err
	}
	log.Info(ctx, "successfully updated task state to completed", logData)
	return nil
}

// Revert handles the revert operations for a dataset version task.
func (e *DatasetVersionTaskExecutor) Revert(ctx context.Context, task *domain.Task) error {
	logData := log.Data{"task_id": task.ID, "job_number": task.JobNumber}

	log.Info(ctx, "starting reversion for dataset version task", logData)

	err := e.jobService.UpdateTaskState(ctx, task.ID, domain.StateCancelled)
	if err != nil {
		log.Error(ctx, "failed to update migration task", err, logData)
		return err
	}
	log.Info(ctx, "completed reversion for dataset version task", logData)
	return nil
}
