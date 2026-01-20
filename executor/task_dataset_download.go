package executor

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dis-migration-service/mapper"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetSDK "github.com/ONSdigital/dp-dataset-api/sdk"
	uploadSDK "github.com/ONSdigital/dp-upload-service/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// DatasetDownloadTaskExecutor executes migration tasks for dataset downloads.
type DatasetDownloadTaskExecutor struct {
	jobService       application.JobService
	clientList       *clients.ClientList
	serviceAuthToken string
}

// NewDatasetDownloadTaskExecutor creates a new DatasetDownloadTaskExecutor
func NewDatasetDownloadTaskExecutor(jobService application.JobService, clientList *clients.ClientList, serviceAuthToken string) *DatasetDownloadTaskExecutor {
	return &DatasetDownloadTaskExecutor{
		jobService:       jobService,
		clientList:       clientList,
		serviceAuthToken: serviceAuthToken,
	}
}

// Migrate handles the migration operations for a dataset download task.
func (e *DatasetDownloadTaskExecutor) Migrate(ctx context.Context, task *domain.Task) error {
	logData := log.Data{"task_id": task.ID, "job_number": task.JobNumber, "source_id": task.Source.ID}

	log.Info(ctx, "starting migration for dataset download task", logData)

	if task == nil || task.Source == nil || task.Target == nil || task.Source.ID == "" {
		err := appErrors.ErrInvalidTask
		log.Error(ctx, "invalid task or missing source/target information", err, logData)
		return err
	}

	fileSize, err := e.clientList.Zebedee.GetFileSize(ctx, e.serviceAuthToken, zebedee.EmptyCollectionId, zebedee.EnglishLangCode, task.Source.ID)
	if err != nil {
		log.Error(ctx, "failed to get file size from zebedee", err, logData)
		return err
	}

	resourceStream, err := e.clientList.Zebedee.GetResourceStream(ctx, e.serviceAuthToken, zebedee.EmptyCollectionId, zebedee.EnglishLangCode, task.Source.ID)
	if err != nil {
		log.Error(ctx, "failed to get file stream from zebedee", err, logData)
		return err
	}
	defer resourceStream.Close()

	uploadMetadata, err := mapper.MapResourceToUploadServiceMetadata(task.Source.ID, fileSize)
	if err != nil {
		log.Error(ctx, "failed to map upload service metadata", err, logData)
		return err
	}

	headers := uploadSDK.Headers{
		ServiceAuthToken: e.serviceAuthToken,
	}

	err = e.clientList.UploadService.Upload(ctx, resourceStream, uploadMetadata, headers)
	if err != nil {
		log.Error(ctx, "failed to upload file to upload service", err, logData)
		return appErrors.ErrFailedToUploadFileToUploadService
	}

	distribution, err := mapper.MapUploadServiceMetadataToDistribution(uploadMetadata)
	if err != nil {
		log.Error(ctx, "failed to map upload service metadata to dataset distribution", err, logData)
		return err
	}

	log.Info(ctx, "updating dataset version metadata with new distribution", logData)
	err = e.updateDownloadMetadata(ctx, task, distribution)
	if err != nil {
		log.Error(ctx, "failed to update dataset version metadata with new distribution", err, logData)
		return err
	}

	err = e.jobService.UpdateTaskState(ctx, task.ID, domain.StateInReview)
	if err != nil {
		log.Error(ctx, "failed to update migration task", err, logData)
		return err
	}
	log.Info(ctx, "completed migration for dataset download task", logData)
	return nil
}

// Publish handles the publish operations for a dataset download task.
func (e *DatasetDownloadTaskExecutor) Publish(ctx context.Context, task *domain.Task) error {
	// Implementation of publish for dataset download
	return nil
}

// PostPublish handles the post-publish operations for a dataset download task.
func (e *DatasetDownloadTaskExecutor) PostPublish(ctx context.Context, task *domain.Task) error {
	// Implementation of post-publish for dataset download
	return nil
}

// Revert handles the revert operations for a dataset download task.
func (e *DatasetDownloadTaskExecutor) Revert(ctx context.Context, task *domain.Task) error {
	// Implementation of revert for dataset download
	return nil
}

func (e *DatasetDownloadTaskExecutor) updateDownloadMetadata(ctx context.Context, task *domain.Task, distribution datasetModels.Distribution) error {
	headers := datasetSDK.Headers{
		AccessToken: e.serviceAuthToken,
	}

	//TODO: use the eTag here to prevent collisions. SDK needs to support it first.
	currentVersion, _, err := e.clientList.DatasetAPI.GetVersionWithHeaders(ctx, headers, task.Target.DatasetID, task.Target.EditionID, task.Target.VersionID)
	if err != nil {
		return err
	}

	*currentVersion.Distributions = append(*currentVersion.Distributions, distribution)

	_, err = e.clientList.DatasetAPI.PutVersion(ctx, headers, task.Target.DatasetID, task.Target.EditionID, task.Target.VersionID, currentVersion)
	if err != nil {
		return err
	}

	return nil
}
