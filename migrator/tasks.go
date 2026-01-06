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
			failErr := mig.failTask(ctx, task)
			if failErr != nil {
				log.Error(ctx, "failed to mark task as failed after failing to get executor", failErr, logData)
				mig.notifyTaskExecutorMissing(ctx, task, err, failErr)
			}
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
			failErr := mig.failTask(ctx, task)
			if failErr != nil {
				log.Error(ctx, "failed to mark task as failed after execution error", failErr, logData)
				mig.notifyTaskExecutionFailure(ctx, task, err, failErr)
				return
			}

			failErr = mig.failJobByJobNumber(ctx, task.JobNumber)
			if failErr != nil {
				log.Error(ctx, "failed to mark job as failed after task execution error", failErr, logData)
				mig.notifyJobFailureAfterTaskError(ctx, task, err, failErr)
				return
			}
		}

		// Success: Check if all tasks are complete and update job state if needed
		checkErr := mig.TriggerJobStateTransitions(ctx, task.JobNumber)
		if checkErr != nil {
			log.Error(ctx, "error checking job state transition", checkErr, logData)
			// Log but don't fail - the job can be checked again later
		}
	}()
}

func (mig *migrator) failTask(ctx context.Context, task *domain.Task) error {
	logData := log.Data{"task_id": task.ID, "job_number": task.JobNumber, "task_state": task.State}

	failureState, err := domain.GetFailureStateForJobState(task.State)
	if err != nil {
		log.Error(ctx, "failed to get failure state for task state", err, logData)
		return err
	}

	logData["failure_state"] = failureState

	err = mig.jobService.UpdateTaskState(ctx, task.ID, failureState)
	if err != nil {
		log.Error(ctx, "failed to update task state to failed", err, logData)
		return err
	}
	return nil
}

// notifyTaskExecutionFailure sends a Slack alarm
// when a task fails to be marked as failed
func (mig *migrator) notifyTaskExecutionFailure(
	ctx context.Context,
	task *domain.Task,
	executionErr error,
	failErr error,
) {
	// Get the job to include details in the notification
	job, err := mig.jobService.GetJob(ctx, task.JobNumber)
	if err != nil {
		log.Error(ctx, "failed to get job for slack notification", err, log.Data{"jobNumber": task.JobNumber})
		return
	}

	details := map[string]interface{}{
		"Task ID":         task.ID,
		"Task Type":       string(task.Type),
		"Job Number":      task.JobNumber,
		"Job Label":       job.Label,
		"Execution Error": executionErr.Error(),
		"Update Error":    failErr.Error(),
	}

	summary := "Task execution failed and task state update failed"

	if err := mig.slackClient.SendAlarm(ctx, summary, executionErr, details); err != nil {
		log.Error(ctx, "failed to send slack alarm for task execution failure", err, log.Data{
			"taskID":    task.ID,
			"jobNumber": task.JobNumber,
		})
	}
}

// notifyJobFailureAfterTaskError sends a Slack alarm
// when a job fails to be marked as failed after a task error
func (mig *migrator) notifyJobFailureAfterTaskError(
	ctx context.Context,
	task *domain.Task,
	executionErr error,
	jobFailErr error,
) {
	// Get the job to include details in the notification
	job, err := mig.jobService.GetJob(ctx, task.JobNumber)
	if err != nil {
		log.Error(ctx, "failed to get job for slack notification", err, log.Data{"jobNumber": task.JobNumber})
		return
	}

	details := map[string]interface{}{
		"Task ID":          task.ID,
		"Task Type":        string(task.Type),
		"Job Number":       task.JobNumber,
		"Job Label":        job.Label,
		"Job State":        string(job.State),
		"Task Error":       executionErr.Error(),
		"Job Update Error": jobFailErr.Error(),
	}

	summary := "Job failed to update to failed state after task execution error"

	if err := mig.slackClient.SendAlarm(ctx, summary, executionErr, details); err != nil {
		log.Error(ctx, "failed to send slack alarm for job failure after task error", err, log.Data{
			"taskID":    task.ID,
			"jobNumber": task.JobNumber,
		})
	}
}

// notifyTaskExecutorMissing sends a Slack alarm when a task executor
// cannot be found AND the task fails to be marked as failed
func (mig *migrator) notifyTaskExecutorMissing(
	ctx context.Context,
	task *domain.Task,
	executorErr error,
	failErr error,
) {
	// Get the job to include details in the notification
	job, err := mig.jobService.GetJob(ctx, task.JobNumber)
	if err != nil {
		log.Error(ctx, "failed to get job for slack notification", err, log.Data{"jobNumber": task.JobNumber})
		return
	}

	details := map[string]interface{}{
		"Task ID":        task.ID,
		"Task Type":      string(task.Type),
		"Task State":     string(task.State),
		"Job Number":     task.JobNumber,
		"Job Label":      job.Label,
		"Executor Error": executorErr.Error(),
		"Update Error":   failErr.Error(),
	}

	summary := "Task executor not found and task state update failed"

	if err := mig.slackClient.SendAlarm(ctx, summary, executorErr, details); err != nil {
		log.Error(ctx, "failed to send slack alarm for missing task executor", err, log.Data{
			"taskID":    task.ID,
			"jobNumber": task.JobNumber,
		})
	}
}

// notifyTaskExecutorMissingWarning sends a Slack warning when a task executor
// cannot be found (but task was successfully marked as failed)
func (mig *migrator) notifyTaskExecutorMissingWarning(
	ctx context.Context,
	task *domain.Task,
	executorErr error,
) {
	// Get the job to include details in the notification
	job, err := mig.jobService.GetJob(ctx, task.JobNumber)
	if err != nil {
		log.Error(ctx, "failed to get job for slack notification", err, log.Data{"jobNumber": task.JobNumber})
		return
	}

	details := map[string]interface{}{
		"Task ID":    task.ID,
		"Task Type":  string(task.Type),
		"Task State": string(task.State),
		"Job Number": task.JobNumber,
		"Job Label":  job.Label,
		"Error":      executorErr.Error(),
	}

	summary := "Task executor not found - check migrator configuration"

	if err := mig.slackClient.SendWarning(ctx, summary, details); err != nil {
		log.Error(ctx, "failed to send slack warning for missing task executor", err, log.Data{
			"taskID":    task.ID,
			"jobNumber": task.JobNumber,
		})
	}
}

// notifyUnsupportedTaskState sends a Slack warning
// when a task is in an unsupported state
func (mig *migrator) notifyUnsupportedTaskState(
	ctx context.Context,
	task *domain.Task,
) {
	// Get the job to include details in the notification
	job, err := mig.jobService.GetJob(ctx, task.JobNumber)
	if err != nil {
		log.Error(ctx, "failed to get job for slack notification", err, log.Data{"jobNumber": task.JobNumber})
		return
	}

	details := map[string]interface{}{
		"Task ID":    task.ID,
		"Task Type":  string(task.Type),
		"Task State": string(task.State),
		"Job Number": task.JobNumber,
		"Job Label":  job.Label,
	}

	summary := "Task in unsupported state for execution - check state machine configuration"

	if err := mig.slackClient.SendWarning(ctx, summary, details); err != nil {
		log.Error(ctx, "failed to send slack warning for unsupported task state", err, log.Data{
			"taskID":    task.ID,
			"jobNumber": task.JobNumber,
		})
	}
}
