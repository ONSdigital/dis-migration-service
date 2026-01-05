package migrator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/executor"
	"github.com/ONSdigital/log.go/v2/log"
)

var getJobExecutors = func(jobService application.JobService, appClients *clients.ClientList) map[domain.JobType]executor.JobExecutor {
	jobExecutors := make(map[domain.JobType]executor.JobExecutor)
	jobExecutors[domain.JobTypeStaticDataset] = executor.NewStaticDatasetJobExecutor(jobService, appClients)
	return jobExecutors
}

func (mig *migrator) getJobExecutor(ctx context.Context, job *domain.Job) (executor.JobExecutor, error) {
	jobExecutor := mig.jobExecutors[job.Config.Type]
	if jobExecutor == nil {
		return nil, fmt.Errorf("no executor found for task type: %s", job.Config.Type)
	}
	return jobExecutor, nil
}

func (mig *migrator) monitorJobs(ctx context.Context) {
	log.Info(ctx, "monitoring jobs", log.Data{"poll_interval": mig.pollInterval})

	for {
		select {
		case <-ctx.Done():
			log.Info(ctx, "stopping monitoring jobs")
			return
		default:
			job, err := mig.jobService.ClaimJob(ctx)
			if err != nil && !errors.Is(err, context.Canceled) {
				log.Error(ctx, "error claiming job", err)
				time.Sleep(mig.pollInterval)
				continue
			}
			if job == nil {
				select {
				case <-ctx.Done():
					log.Info(ctx, "stopping monitoring jobs")
					return
				case <-time.After(mig.pollInterval):
					continue
				}
			}

			log.Info(ctx, "claimed job", log.Data{"job_id": job.ID, "job_state": job.State})
			mig.executeJob(ctx, job)
		}
	}
}

// executeJob executes the provided job in a separate goroutine based on
// it's state
func (mig *migrator) executeJob(ctx context.Context, job *domain.Job) {
	mig.wg.Add(1)
	go func() {
		defer mig.wg.Done()

		logData := log.Data{"job_id": job.ID, "job_state": job.State}

		select {
		case mig.semaphore <- struct{}{}:
			defer func() { <-mig.semaphore }()
		case <-ctx.Done():
			return
		}

		jobExecutor, err := mig.getJobExecutor(ctx, job)
		if err != nil {
			log.Error(ctx, "failed to get job executor", err, logData)
			failErr := mig.failJob(ctx, job)
			if failErr != nil {
				log.Error(ctx, "failed to mark job as failed after failing to get executor", failErr, logData)
				mig.notifyJobExecutorMissing(ctx, job, err, failErr)
			}
			return
		}

		switch job.State {
		case domain.StateMigrating:
			err = jobExecutor.Migrate(ctx, job)
		default:
			err = fmt.Errorf("unsupported job state: %s", job.State)
			log.Error(ctx, "unsupported job state for execution", err, logData)
		}

		if err != nil {
			log.Error(ctx, "error executing job", err, logData)
			failErr := mig.failJob(ctx, job)
			if failErr != nil {
				log.Error(ctx, "failed to mark job as failed after execution error", failErr, logData)
				mig.notifyJobExecutionFailure(ctx, job, err, failErr)
			}
		}
	}()
}

func (mig *migrator) failJobByJobNumber(ctx context.Context, jobNumber int) error {
	job, err := mig.jobService.GetJob(ctx, jobNumber)
	if err != nil {
		log.Error(ctx, "failed to get job by job number to fail it", err)
		return err
	}

	return mig.failJob(ctx, job)
}

func (mig *migrator) failJob(ctx context.Context, job *domain.Job) error {
	logData := log.Data{"job_id": job.ID, "job_state": job.State}

	if domain.IsFailedState(job.State) {
		log.Info(ctx, "job is already in a failed state, skipping fail operation", logData)
		return nil
	}

	failureState, err := domain.GetFailureStateForJobState(job.State)
	if err != nil {
		log.Error(ctx, "failed to get failure state for job state", err, logData)
		return err
	}

	logData["failure_state"] = failureState
	err = mig.jobService.UpdateJobState(ctx, job.JobNumber, failureState, "")
	if err != nil {
		log.Error(ctx, "failed to update task state to failed", err, logData)
		return err
	}

	return nil
}

// notifyJobExecutionFailure sends a Slack alarm when a job fails to execute
// AND fails to be marked as failed
func (mig *migrator) notifyJobExecutionFailure(
	ctx context.Context,
	job *domain.Job,
	executionErr error,
	failErr error,
) {
	details := map[string]interface{}{
		"Job ID":          job.ID,
		"Job Number":      job.JobNumber,
		"Job Label":       job.Label,
		"Job State":       string(job.State),
		"Execution Error": executionErr.Error(),
		"Update Error":    failErr.Error(),
	}

	summary := "Job execution failed and job state update failed"

	if err := mig.slackClient.SendAlarm(ctx, summary, executionErr, details); err != nil {
		log.Error(ctx, "failed to send slack alarm for job execution failure", err, log.Data{
			"jobID":     job.ID,
			"jobNumber": job.JobNumber,
		})
	}
}

// notifyJobExecutorMissing sends a Slack alarm when a job executor
// cannot be found AND the job fails to be marked as failed
func (mig *migrator) notifyJobExecutorMissing(
	ctx context.Context,
	job *domain.Job,
	executorErr error,
	failErr error,
) {
	details := map[string]interface{}{
		"Job ID":         job.ID,
		"Job Number":     job.JobNumber,
		"Job Label":      job.Label,
		"Job State":      string(job.State),
		"Job Type":       string(job.Config.Type),
		"Executor Error": executorErr.Error(),
		"Update Error":   failErr.Error(),
	}

	summary := "Job executor not found and job state update failed"

	if err := mig.slackClient.SendAlarm(ctx, summary, executorErr, details); err != nil {
		log.Error(ctx, "failed to send slack alarm for missing job executor", err, log.Data{
			"jobID":     job.ID,
			"jobNumber": job.JobNumber,
		})
	}
}

// notifyJobExecutorMissingWarning sends a Slack warning when a job executor
// cannot be found (but job was successfully marked as failed)
func (mig *migrator) notifyJobExecutorMissingWarning(
	ctx context.Context,
	job *domain.Job,
	executorErr error,
) {
	details := map[string]interface{}{
		"Job ID":     job.ID,
		"Job Number": job.JobNumber,
		"Job Label":  job.Label,
		"Job State":  string(job.State),
		"Job Type":   string(job.Config.Type),
		"Error":      executorErr.Error(),
	}

	summary := "Job executor not found - check migrator configuration"

	if err := mig.slackClient.SendWarning(ctx, summary, details); err != nil {
		log.Error(ctx, "failed to send slack warning for missing job executor", err, log.Data{
			"jobID":     job.ID,
			"jobNumber": job.JobNumber,
		})
	}
}

// notifyUnsupportedJobState sends a Slack warning when a job is in an unsupported state
func (mig *migrator) notifyUnsupportedJobState(
	ctx context.Context,
	job *domain.Job,
) {
	details := map[string]interface{}{
		"Job ID":     job.ID,
		"Job Number": job.JobNumber,
		"Job Label":  job.Label,
		"Job State":  string(job.State),
		"Job Type":   string(job.Config.Type),
	}

	summary := "Job in unsupported state for execution - check state machine configuration"

	if err := mig.slackClient.SendWarning(ctx, summary, details); err != nil {
		log.Error(ctx, "failed to send slack warning for unsupported job state", err, log.Data{
			"jobID":     job.ID,
			"jobNumber": job.JobNumber,
		})
	}
}

// notifyJobStateCompletion sends a Slack notification when a job
// completes an active processing state
func (mig *migrator) notifyJobStateCompletion(
	ctx context.Context,
	job *domain.Job,
	newState domain.State,
) {
	details := map[string]interface{}{
		"Job Label":  job.Label,
		"Job Number": job.JobNumber,
		"Job State":  string(newState),
		"Job ID":     job.ID,
	}

	summary := getJobCompletionSummary(job.State, newState)

	if err := mig.slackClient.SendInfo(ctx, summary, details); err != nil {
		log.Error(
			ctx,
			"failed to send slack notification for job state completion",
			err,
			log.Data{
				"jobID":    job.ID,
				"jobState": newState,
			},
		)
		// Don't return error - notification failure shouldn't stop processing
	}
}

// getJobCompletionSummary returns a human-readable summary of
// the state completion
func getJobCompletionSummary(fromState, toState domain.State) string {
	switch fromState {
	case domain.StateMigrating:
		return "Job migration completed successfully"
	case domain.StatePublishing:
		return "Job publishing completed successfully"
	case domain.StatePostPublishing:
		return "Job post-publishing completed successfully"
	default:
		return "Job state updated"
	}
}
