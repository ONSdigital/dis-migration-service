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
}

// NewDefaultMigrator creates a new default migrator with the
// provided job service and clients
func NewDefaultMigrator(cfg *config.Config, jobService application.JobService, appClients *clients.ClientList) *migrator {
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
	go mig.monitorJobs(ctx)
	go mig.monitorTasks(ctx)
}

// Shutdown waits for all ongoing migrations to complete or times out
func (mig *migrator) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		mig.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		err := fmt.Errorf("timed out waiting for background tasks to complete")
		log.Error(ctx, "error shutting down migrator", err)
		return err
	}
}
