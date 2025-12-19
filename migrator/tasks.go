package migrator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/executor"
	"github.com/ONSdigital/log.go/v2/log"
)

var getTaskExecutors = func(jobService application.JobService, appClients *clients.ClientList, cfg *config.Config) map[domain.TaskType]executor.TaskExecutor {
	taskExecutors := make(map[domain.TaskType]executor.TaskExecutor)
	taskExecutors[domain.TaskTypeDatasetSeries] = executor.NewDatasetSeriesTaskExecutor(jobService, appClients, cfg.ServiceAuthToken)
	taskExecutors[domain.TaskTypeDatasetEdition] = executor.NewDatasetEditionTaskExecutor(jobService, appClients)
	return taskExecutors
}

func (mig *migrator) getTaskExecutor(ctx context.Context, task *domain.Task) (executor.TaskExecutor, error) {
	taskExecutor := mig.taskExecutors[task.Type]
	if taskExecutor == nil {
		return nil, fmt.Errorf("no executor found for task type: %s", task.Type)
	}
	return taskExecutor, nil
}

func (mig *migrator) monitorTasks(ctx context.Context) {
	log.Info(ctx, "monitoring tasks", log.Data{"pollInterval": mig.pollInterval})

	for {
		select {
		case <-ctx.Done():
			log.Info(ctx, "stopping monitoring tasks")
			return
		default:
			task, err := mig.jobService.ClaimTask(ctx)
			if err != nil && !errors.Is(err, context.Canceled) {
				log.Error(ctx, "error claiming task", err)
				time.Sleep(mig.pollInterval)
				continue
			}
			if task == nil {
				select {
				case <-ctx.Done():
					log.Info(ctx, "stopping monitoring tasks")
					return
				case <-time.After(mig.pollInterval):
					continue
				}
			}
			log.Info(ctx, "claimed task", log.Data{"taskID": task.ID, "taskState": task.State})
			mig.executeTask(ctx, task)
		}
	}
}

// executeTask executes a task based on its state
func (mig *migrator) executeTask(ctx context.Context, task *domain.Task) {
	mig.wg.Add(1)
	go func() {
		defer mig.wg.Done()

		select {
		case mig.semaphore <- struct{}{}:
			defer func() { <-mig.semaphore }()
		case <-ctx.Done():
			return
		}

		taskExecutor, err := mig.getTaskExecutor(ctx, task)
		if err != nil {
			log.Error(ctx, "failed to get task executor", err, log.Data{"task": task.ID, "jobNumber": task.JobNumber, "taskType": task.Type})
			failErr := mig.failTask(ctx, task)
			if failErr != nil {
				log.Error(ctx, "failed to mark task as failed after failing to get executor", failErr, log.Data{"taskID": task.ID, "taskState": task.State})
			}
			return
		}

		// err is left hanging here for the catch-all error handler below as the handling is the same for all task states
		switch task.State {
		case domain.StateMigrating:
			err = taskExecutor.Migrate(ctx, task)
		default:
			err = fmt.Errorf("unsupported task state: %s", task.State)
			log.Error(ctx, "unsupported task state for execution", err, log.Data{"taskID": task.ID, "taskState": task.State})
		}

		if err != nil {
			log.Error(ctx, "error executing task", err, log.Data{"taskID": task.ID, "taskState": task.State})
			failErr := mig.failTask(ctx, task)
			if failErr != nil {
				//TODO: flag this in slack.
				log.Error(ctx, "failed to mark task as failed after execution error", failErr, log.Data{"taskID": task.ID, "taskState": task.State})
			}

			failErr = mig.failJobByJobNumber(ctx, task.JobNumber)
			if failErr != nil {
				//TODO: flag this in slack.
				log.Error(ctx, "failed to mark job as failed after task execution error", failErr, log.Data{"taskID": task.ID, "taskState": task.State, "jobNumber": task.JobNumber})
			}
		}

		// Success: Check if all tasks are complete and update job state if needed
		checkErr := mig.TriggerJobStateTransitions(ctx, task.JobID)
		if checkErr != nil {
			log.Error(ctx, "error checking job state transition", checkErr, log.Data{
				"taskID": task.ID, "jobID": task.JobID,
			})
			// Log but don't fail - the job can be checked again later
		}
	}()
}

func (mig *migrator) failTask(ctx context.Context, task *domain.Task) error {
	logData := log.Data{"taskID": task.ID, "jobNumber": task.JobNumber, "taskState": task.State}

	failureState, err := domain.GetFailureStateForJobState(task.State)
	if err != nil {
		log.Error(ctx, "failed to get failure state for task state", err, log.Data{"taskState": task.State})
		return err
	}

	logData["failureState"] = failureState

	err = mig.jobService.UpdateTaskState(ctx, task.ID, failureState)
	if err != nil {
		log.Error(ctx, "failed to update task state to failed", err, logData)
		return err
	}

	return nil
}
