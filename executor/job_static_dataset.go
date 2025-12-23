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
	clientList *clients.ClientList
	jobService application.JobService
}

// NewStaticDatasetJobExecutor creates a new StaticDatasetJobExecutor
func NewStaticDatasetJobExecutor(jobService application.JobService, clientList *clients.ClientList) *StaticDatasetJobExecutor {
	return &StaticDatasetJobExecutor{
		jobService: jobService,
		clientList: clientList,
	}
}

// Migrate handles the migration operations for a static dataset job.
func (e *StaticDatasetJobExecutor) Migrate(ctx context.Context, job *domain.Job) error {
	log.Info(ctx, "starting migration for job", log.Data{"jobID": job.ID})

	datasetSeriesTask := domain.NewTask(job.JobNumber)

	datasetSeriesTask.Type = domain.TaskTypeDatasetSeries
	datasetSeriesTask.Source = &domain.TaskMetadata{
		ID: job.Config.SourceID,
	}

	datasetSeriesTask.Target = &domain.TaskMetadata{
		ID: job.Config.TargetID,
	}

	_, err := e.jobService.CreateTask(ctx, job.JobNumber, &datasetSeriesTask)
	if err != nil {
		log.Error(ctx, "failed to create migration task", err, log.Data{"jobID": job.ID, "task": datasetSeriesTask})
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
