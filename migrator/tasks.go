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
	"github.com/ONSdigital/dis-migration-service/slack"
	"github.com/ONSdigital/log.go/v2/log"
)

var getTaskExecutors = func(jobService application.JobService, appClients *clients.ClientList, cfg *config.Config) map[domain.TaskType]executor.TaskExecutor {
	taskExecutors := make(map[domain.TaskType]executor.TaskExecutor)
	taskExecutors[domain.TaskTypeDatasetSeries] = executor.NewDatasetSeriesTaskExecutor(jobService, appClients, cfg.ServiceAuthToken)
	taskExecutors[domain.TaskTypeDatasetEdition] = executor.NewDatasetEditionTaskExecutor(jobService, appClients, cfg.ServiceAuthToken)
	taskExecutors[domain.TaskTypeDatasetVersion] = executor.NewDatasetVersionTaskExecutor(jobService, appClients, cfg.ServiceAuthToken)
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
	log.Info(ctx, "monitoring tasks", log.Data{"poll_interval": mig.pollInterval})

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
			log.Info(ctx, "claimed task", log.Data{"task_id": task.ID, "task_state": task.State})
			mig.executeTask(ctx, task)
		}
	}
}

// executeTask executes a task based on its state
func (mig *migrator) executeTask(ctx context.Context, task *domain.Task) {
	mig.wg.Add(1)
	go func() {
		defer mig.wg.Done()

		logData := log.Data{"task_id": task.ID, "job_number": task.JobNumber, "task_type": task.Type}

		select {
		case mig.semaphore <- struct{}{}:
			defer func() { <-mig.semaphore }()
		case <-ctx.Done():
			return
		}

		taskExecutor, err := mig.getTaskExecutor(ctx, task)
		if err != nil {
			log.Error(ctx, "failed to get task executor", err, logData)
			_ = mig.failTask(ctx, task, err, failureReasonExecutorMissing)
			return
		}

		// err is left hanging here for the catch-all error handler below as the handling is the same for all task states
		switch task.State {
		case domain.StateMigrating:
			err = taskExecutor.Migrate(ctx, task)
		default:
			err = fmt.Errorf("unsupported task state: %s", task.State)
			log.Error(ctx, "unsupported task state for execution", err, logData)
		}

		if err != nil {
			log.Error(ctx, "error executing task", err, logData)
			failErr := mig.failTask(ctx, task, err, failureReasonExecutionFailed)
			if failErr != nil {
				return
			}

			failErr = mig.failJobByJobNumber(ctx, task.JobNumber, err, failureReasonExecutionFailed)
			if failErr != nil {
				log.Error(ctx, "failed to fail job after task execution error", failErr, logData)
				return
			}
			return
		}

		// Success: Check if all tasks are complete and update job state if needed
		checkErr := mig.TriggerJobStateTransitions(ctx, task.JobNumber)
		if checkErr != nil {
			log.Error(ctx, "error checking job state transition", checkErr, logData)
			// Log but don't fail - the job can be checked again later
		}
	}()
}

func (mig *migrator) failTask(ctx context.Context, task *domain.Task, originalErr error, failureReason string) error {
	logData := log.Data{"task_id": task.ID, "job_number": task.JobNumber, "task_state": task.State}
	slackDetails := slack.SlackDetails{
		"Task ID":        task.ID,
		"Job Number":     task.JobNumber,
		"Task State":     task.State,
		"Failure Reason": failureReason,
	}

	failureState, err := domain.GetFailureStateForJobState(task.State)
	if err != nil {
		log.Error(ctx, "failed to get failure state for task state", err, logData)
		return err
	}

	logData["failure_state"] = failureState

	err = mig.jobService.UpdateTaskState(ctx, task.ID, failureState)
	if err != nil {
		log.Error(ctx, "failed to update task state to failed", err, logData)

		slackDetails["Failure State"] = failureState
		slackDetails["Original Error"] = originalErr.Error()
		slackDetails["Update Error"] = err.Error()

		slackErr := mig.slackClient.SendAlarm(ctx, EventUpdateTaskStateFailed, nil, slackDetails)
		if slackErr != nil {
			log.Error(ctx, "failed to send slack notification", slackErr, logData)
		}
		return err
	}

	return nil
}
