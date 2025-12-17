package executor

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/log.go/v2/log"
)

// DatasetSeriesTaskExecutor executes migration tasks for dataset series.
type DatasetSeriesTaskExecutor struct {
	jobService application.JobService
	clientList *clients.ClientList
}

// NewDatasetSeriesTaskExecutor creates a new DatasetSeriesTaskExecutor
func NewDatasetSeriesTaskExecutor(jobService application.JobService, clientList *clients.ClientList) *DatasetSeriesTaskExecutor {
	return &DatasetSeriesTaskExecutor{
		jobService: jobService,
		clientList: clientList,
	}
}

// Migrate handles the migration operations for a dataset series task.
func (e *DatasetSeriesTaskExecutor) Migrate(ctx context.Context, task *domain.Task) error {
	logData := log.Data{"taskID": task.ID, "jobID": task.JobNumber}

	log.Info(ctx, "starting migration for dataset series task", logData)

	//TODO: add dataset series migration logic

	err := e.jobService.UpdateTaskState(ctx, task.ID, domain.StateInReview)
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
