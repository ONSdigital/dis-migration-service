package executor

import (
	"context"
	"strings"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/log.go/v2/log"
)

// DatasetEditionTaskExecutor executes migration tasks for dataset editions.
type DatasetEditionTaskExecutor struct {
	jobService application.JobService
	clientList *clients.ClientList
}

// NewDatasetEditionTaskExecutor creates a new DatasetEditionTaskExecutor
func NewDatasetEditionTaskExecutor(jobService application.JobService, clientList *clients.ClientList) *DatasetEditionTaskExecutor {
	return &DatasetEditionTaskExecutor{
		jobService: jobService,
		clientList: clientList,
	}
}

// Migrate handles the migration operations for a dataset edition task.
func (e *DatasetEditionTaskExecutor) Migrate(ctx context.Context, task *domain.Task) error {
	logData := log.Data{"taskID": task.ID, "jobID": task.JobID}

	log.Info(ctx, "starting migration for dataset edition task", logData)

	sourceData, err := e.clientList.Zebedee.GetDataset(ctx, "", "", "en", task.Source.ID)
	if err != nil {
		log.Error(ctx, "failed to get source edition data from zebedee", err, logData)
		return err
	}

	editionID := extractLastSegmentFromURI(sourceData.URI)

	logData["editionID"] = editionID

	task.Target.ID = editionID

	err = e.jobService.UpdateTask(ctx, task)
	if err != nil {
		log.Error(ctx, "failed to update edition migration task with target id", err, logData)
		return err
	}

	currentVersionTask := domain.NewTask(task.JobID)
	currentVersionTask.Type = domain.TaskTypeDatasetVersion
	currentVersionTask.Source = &domain.TaskMetadata{
		ID: sourceData.URI,
	}
	currentVersionTask.Target = &domain.TaskMetadata{
		DatasetID: task.Target.DatasetID,
		EditionID: editionID,
	}

	_, err = e.jobService.CreateTask(ctx, task.JobID, &currentVersionTask)
	if err != nil {
		logData["versionURI"] = currentVersionTask.Source.ID
		log.Error(ctx, "failed to create migration version task for edition", err, logData)
		return err
	}

	for _, previousVersion := range sourceData.Versions {
		versionTask := domain.NewTask(task.JobID)

		versionTask.Type = domain.TaskTypeDatasetVersion
		versionTask.Source = &domain.TaskMetadata{
			ID: previousVersion.URI,
		}
		versionTask.Target = &domain.TaskMetadata{
			DatasetID: task.Target.DatasetID,
			EditionID: editionID,
		}

		_, err := e.jobService.CreateTask(ctx, task.JobID, &versionTask)
		if err != nil {
			logData["versionURI"] = versionTask.Source.ID
			log.Error(ctx, "failed to create migration task for previous version of edition", err, logData)
			return err
		}
	}

	err = e.jobService.UpdateTaskState(ctx, task.ID, domain.TaskStateInReview)
	if err != nil {
		log.Error(ctx, "failed to update migration task", err, logData)
		return err
	}
	log.Info(ctx, "completed migration for dataset edition task", logData)
	return nil
}

// Publish handles the publish operations for a dataset edition task.
func (e *DatasetEditionTaskExecutor) Publish(ctx context.Context, task *domain.Task) error {
	// Implementation of publish for dataset edition
	return nil
}

// PostPublish handles the post-publish operations for a dataset edition task.
func (e *DatasetEditionTaskExecutor) PostPublish(ctx context.Context, task *domain.Task) error {
	// Implementation of post-publish for dataset edition
	return nil
}

// Revert handles the revert operations for a dataset edition task.
func (e *DatasetEditionTaskExecutor) Revert(ctx context.Context, task *domain.Task) error {
	// Implementation of revert for dataset edition
	return nil
}

func extractLastSegmentFromURI(uri string) string {
	parts := strings.Split(uri, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return ""
}
