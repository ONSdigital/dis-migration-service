package migrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/executor"
	"github.com/ONSdigital/log.go/v2/log"
)

type migrator struct {
	jobService    application.JobService
	jobExecutors  map[domain.JobType]executor.JobExecutor
	taskExecutors map[domain.TaskType]executor.TaskExecutor
	wg            sync.WaitGroup
	semaphore     chan struct{}
	pollInterval  time.Duration
	stopJobsFunc  context.CancelFunc
}

// NewDefaultMigrator creates a new default migrator with the
// provided job service and clients
func NewDefaultMigrator(cfg *config.Config, jobService application.JobService, appClients *clients.ClientList) *migrator {
	// Calling GetNextJobNumber to create the JobNumber counter in MongoDB if it does not already exist
	ctx := context.Background()
	_, err := jobService.GetNextJobNumber(ctx)
	if err != nil {
		log.Error(ctx, "error getting or creating job number counter", err)
	}

	jobExecutors := getJobExecutors(jobService, appClients)
	taskExecutors := getTaskExecutors(jobService, appClients)

	return &migrator{
		jobService:    jobService,
		jobExecutors:  jobExecutors,
		taskExecutors: taskExecutors,
		pollInterval:  cfg.MigratorPollInterval,
		// Semaphore to limit concurrent migrations
		semaphore: make(chan struct{}, cfg.MigratorMaxConcurrentExecutions),
	}
}

// Start begins monitoring for jobs and tasks to process
func (mig *migrator) Start(ctx context.Context) {
	log.Info(ctx, "starting migrator")
	ctx, cancel := context.WithCancel(context.Background())
	mig.stopJobsFunc = cancel
	mig.wg.Add(2)
	go func() {
		defer mig.wg.Done()
		mig.monitorJobs(ctx)
	}()
	go func() {
		defer mig.wg.Done()
		mig.monitorTasks(ctx)
	}()
}

// Shutdown waits for all ongoing migrations to complete or times out
func (mig *migrator) Shutdown(ctx context.Context) error {
	log.Info(ctx, "shutting down migrator")

	if mig.stopJobsFunc != nil {
		mig.stopJobsFunc()
	}

	done := make(chan struct{})
	go func() {
		mig.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info(ctx, "migrator shut down completed successfully")
		return nil
	case <-ctx.Done():
		err := fmt.Errorf("timed out waiting for background tasks to complete")
		log.Error(ctx, "error shutting down migrator", err)
		return err
	}
}
