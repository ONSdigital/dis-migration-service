package executor

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/mapper"
	"github.com/ONSdigital/log.go/v2/log"
)

// DatasetSeriesTaskExecutor executes migration tasks for dataset series.
type DatasetSeriesTaskExecutor struct {
	jobService application.JobService
	clientList *clients.ClientList
}

// NewDatasetSeriesTaskExecutor creates a new DatasetSeriesTaskExecutor
func NewDatasetSeriesTaskExecutor(ctx context.Context, jobService application.JobService, clientList *clients.ClientList) *DatasetSeriesTaskExecutor {
	return &DatasetSeriesTaskExecutor{
		jobService: jobService,
		clientList: clientList,
	}
}

// Migrate handles the migration operations for a dataset series task.
func (e *DatasetSeriesTaskExecutor) Migrate(ctx context.Context, task *domain.Task) error {
	logData := log.Data{"taskID": task.ID, "jobID": task.JobID}

	log.Info(ctx, "starting migration for dataset series task", logData)
	sourceData, err := e.clientList.Zebedee.GetDatasetLandingPage(ctx, "", "", "en", task.Source.ID)
	if err != nil {
		log.Error(ctx, "failed to get source dataset data from zebedee", err, logData)
		return err
	}

	task.Source.Label = sourceData.Description.Title

	targetData, err := mapper.MapDatasetLandingPageToDatasetAPI(task.Target.ID, sourceData)
	if err != nil {
		log.Error(ctx, "failed to map dataset landing page to dataset API model", err, logData)
		return err
	}

	//TODO: add to dataset API
	log.Info(ctx, "this is a fake log for creating dataset in dataset API", log.Data{"dataset": targetData})

	for _, edition := range sourceData.Datasets {
		editionTask := domain.NewTask(task.JobID)

		editionTask.Type = domain.TaskTypeDatasetEdition
		editionTask.Source = &domain.TaskMetadata{
			ID: edition.URI,
		}
		editionTask.Target = &domain.TaskMetadata{
			DatasetID: task.Target.ID,
		}

		_, err := e.jobService.CreateTask(ctx, task.JobID, &editionTask)
		if err != nil {
			log.Error(ctx, "failed to create migration task", err, log.Data{"jobID": task.JobID, "task": editionTask})
			return err
		}
	}

	task.State = domain.TaskStateInReview

	err = e.jobService.UpdateTask(ctx, task)
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
