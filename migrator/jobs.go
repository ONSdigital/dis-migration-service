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
	"github.com/ONSdigital/dis-migration-service/slack"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	failureReasonExecutorMissing = "executor_missing"
	failureReasonExecutionFailed = "execution_failed"

	// EventJobFailed is sent when a job fails.
	EventJobFailed = "Migration Job Failed"
	// EventJobCompleted is sent when a job completes successfully.
	EventJobCompleted = "Migration Job Completed"
	// EventUpdateJobStateFailed is sent when updating a job state fails.
	EventUpdateJobStateFailed = "Failed to update job state"
	// EventUpdateTaskStateFailed is sent when updating a task state fails.
	EventUpdateTaskStateFailed = "Failed to update task state"
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
			_ = mig.failJob(ctx, job, err, failureReasonExecutorMissing)
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
			_ = mig.failJob(ctx, job, err, failureReasonExecutionFailed)
		}
	}()
}

func (mig *migrator) failJobByJobNumber(ctx context.Context, jobNumber int, originalErr error, failureReason string) error {
	job, err := mig.jobService.GetJob(ctx, jobNumber)
	if err != nil {
		log.Error(ctx, "failed to get job by job number to fail it", err)
		return err
	}

	return mig.failJob(ctx, job, originalErr, failureReason)
}

func (mig *migrator) failJob(ctx context.Context, job *domain.Job, originalErr error, failureReason string) error {
	logData := log.Data{"job_id": job.ID, "job_state": job.State}

	slackDetails := slack.SlackDetails{
		"Job Number":     job.JobNumber,
		"Job Label":      job.Label,
		"Job State":      job.State,
		"Failure Reason": failureReason,
	}

	err := mig.slackClient.SendInfo(ctx, EventJobFailed, slackDetails)
	if err != nil {
		log.Error(ctx, "failed to send slack notification", err, logData)
	}

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

		// Notify about failure - include both errors in details
		slackDetails["Failure State"] = failureState
		slackDetails["Original Error"] = originalErr.Error()
		slackDetails["Update Error"] = err.Error()

		err = mig.slackClient.SendAlarm(ctx, EventUpdateJobStateFailed, nil, slackDetails)
		if err != nil {
			log.Error(ctx, "failed to send slack notification", err, logData)
		}
		return err
	}
	return nil
}
