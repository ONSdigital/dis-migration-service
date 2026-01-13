package executor

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/mapper"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// DatasetSeriesTaskExecutor executes migration tasks for dataset series.
type DatasetSeriesTaskExecutor struct {
	jobService       application.JobService
	clientList       *clients.ClientList
	serviceAuthToken string
}

// NewDatasetSeriesTaskExecutor creates a new DatasetSeriesTaskExecutor
func NewDatasetSeriesTaskExecutor(jobService application.JobService, clientList *clients.ClientList, serviceAuthToken string) *DatasetSeriesTaskExecutor {
	return &DatasetSeriesTaskExecutor{
		jobService:       jobService,
		clientList:       clientList,
		serviceAuthToken: serviceAuthToken,
	}
}

// Migrate handles the migration operations for a dataset series task.
func (e *DatasetSeriesTaskExecutor) Migrate(ctx context.Context, task *domain.Task) error {
	logData := log.Data{"taskID": task.ID, "jobNumber": task.JobNumber}

	log.Info(ctx, "starting migration for dataset series task", logData)

	sourceData, err := e.clientList.Zebedee.GetDatasetLandingPage(ctx, e.serviceAuthToken, zebedee.EmptyCollectionId, zebedee.EnglishLangCode, task.Source.ID)
	if err != nil {
		log.Error(ctx, "failed to get source dataset series data from zebedee", err, logData)
		return err
	}

	targetData, err := mapper.MapDatasetLandingPageToDatasetAPI(task.Target.ID, sourceData)
	if err != nil {
		log.Error(ctx, "failed to map dataset landing page to dataset API model", err, logData)
		return err
	}

	headers := sdk.Headers{
		AccessToken: e.serviceAuthToken,
	}

	_, err = e.clientList.DatasetAPI.CreateDataset(ctx, headers, *targetData)
	if err != nil {
		log.Error(ctx, "failed to create target dataset in dataset API", err, logData)
		return err
	}

	for _, edition := range sourceData.Datasets {
		editionTask := domain.NewTask(task.JobNumber)

		editionTask.Type = domain.TaskTypeDatasetEdition
		editionTask.Source = &domain.TaskMetadata{
			ID: edition.URI,
		}
		editionTask.Target = &domain.TaskMetadata{
			DatasetID: task.Target.ID,
		}

		_, err := e.jobService.CreateTask(ctx, task.JobNumber, &editionTask)
		if err != nil {
			logData["editionURI"] = edition.URI
			log.Error(ctx, "failed to create migration task for edition", err, logData)
			return err
		}
	}

	err = e.jobService.UpdateTaskState(ctx, task.ID, domain.StateInReview)
	if err != nil {
		log.Error(ctx, "failed to update migration task", err, logData)
		return err
	}
	log.Info(ctx, "completed migration for dataset series task", logData)
	return nil
}

// Publish handles the publish operations for a dataset series task.
func (e *DatasetSeriesTaskExecutor) Publish(ctx context.Context, task *domain.Task) error {
	// Implementation of publish for static dataset
	return nil
}

// PostPublish handles post-publish operations for a dataset series task.
func (e *DatasetSeriesTaskExecutor) PostPublish(ctx context.Context, task *domain.Task) error {
	// Implementation of post-publish for static dataset
	return nil
}

// Revert handles the revert operations for a dataset series task.
func (e *DatasetSeriesTaskExecutor) Revert(ctx context.Context, task *domain.Task) error {
	// Implementation of revert for static dataset
	return nil
}
