package executor

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
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
	return nil
}

// PostPublish handles the post-publish operations for a static dataset job.
func (e *StaticDatasetJobExecutor) PostPublish(ctx context.Context, job *domain.Job) error {
	// Implementation of post-publish for static dataset
	return nil
}

// Revert handles the revert operations for a static dataset job.
func (e *StaticDatasetJobExecutor) Revert(ctx context.Context, job *domain.Job) error {
	// Implementation of revert for static dataset
	return nil
}
