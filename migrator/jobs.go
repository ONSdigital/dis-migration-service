package migrator

import (
	"context"
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
	jobExecutors[domain.JobTypeStaticDataset] = executor.NewStaticDatasetJobExecutor(context.Background(), jobService, appClients)
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
	log.Info(ctx, "monitoring jobs")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			log.Info(ctx, "claiming jobs")

			job, err := mig.jobService.ClaimJob(ctx)
			if err != nil {
				log.Error(ctx, "error claiming job", err)
				continue
			}
			if job == nil {
				// No jobs available, wait before retrying
				log.Info(ctx, "no jobs available to claim, sleeping", log.Data{"pollInterval": mig.pollInterval})

				time.Sleep(mig.pollInterval)
				continue
			}

			log.Info(ctx, "claimed job", log.Data{"jobID": job.ID, "jobState": job.State})
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
		mig.semaphore <- struct{}{}
		defer func() { <-mig.semaphore }()

		jobExecutor, err := mig.getJobExecutor(ctx, job)
		if err != nil {
			log.Error(ctx, "failed to get job executor", err, log.Data{"jobID": job.ID})
			failErr := mig.failJob(ctx, job)
			if failErr != nil {
				log.Error(ctx, "failed to mark job as failed after failing to get executor", failErr, log.Data{"jobID": job.ID, "jobState": job.State})
			}
			return
		}

		switch job.State {
		case domain.JobStateMigrating:
			err = jobExecutor.Migrate(ctx, job)
		default:
			err = fmt.Errorf("unsupported job state: %s", job.State)
			log.Error(ctx, "unsupported job state for execution", err, log.Data{"jobID": job.ID, "jobState": job.State})
		}

		if err != nil {
			log.Error(ctx, "error executing job", err, log.Data{"jobID": job.ID, "jobState": job.State})
			failErr := mig.failJob(ctx, job)
			if failErr != nil {
				// TODO signpost this in slack
				log.Error(ctx, "failed to mark job as failed after execution error", failErr, log.Data{"jobID": job.ID, "jobState": job.State})
			}
		}
	}()
}

func (mig *migrator) failJobByID(ctx context.Context, jobID string) error {
	job, err := mig.jobService.GetJob(ctx, jobID)
	if err != nil {
		log.Error(ctx, "failed to get job by id to fail it", err)
		return err
	}

	return mig.failJob(ctx, job)
}

func (mig *migrator) failJob(ctx context.Context, job *domain.Job) error {
	logData := log.Data{"jobID": job.ID, "jobState": job.State}

	if domain.IsFailedState(job.State) {
		log.Info(ctx, "job is already in a failed state, skipping fail operation", logData)
		return nil
	}

	failureState, err := domain.GetFailureStateForJobState(job.State)
	if err != nil {
		log.Error(ctx, "failed to get failure state for job state", err, log.Data{"jobState": job.State})
		return err
	}

	logData["failureState"] = failureState

	err = mig.jobService.UpdateJobState(ctx, job, failureState)
	if err != nil {
		log.Error(ctx, "failed to update task state to failed", err, logData)
		return err
	}

	return nil
}
