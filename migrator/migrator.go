package migrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/cache"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/executor"
	"github.com/ONSdigital/dis-migration-service/slack"
	"github.com/ONSdigital/log.go/v2/log"
)

type migrator struct {
	jobService    application.JobService
	jobExecutors  map[domain.JobType]executor.JobExecutor
	taskExecutors map[domain.TaskType]executor.TaskExecutor
	slackClient   slack.Clienter
	wg            sync.WaitGroup
	semaphore     chan struct{}
	pollInterval  time.Duration
	stopJobsFunc  context.CancelFunc
	topicCache    *cache.TopicCache
	cfg           *config.Config
	appClients    *clients.ClientList
}

// NewDefaultMigrator creates a new default migrator with the
// provided job service and clients
func NewDefaultMigrator(cfg *config.Config, jobService application.JobService, appClients *clients.ClientList, slackClient slack.Clienter) *migrator {
	jobExecutors := getJobExecutors(jobService, appClients)
	// Task executors will be created when topic cache is set
	taskExecutors := make(map[domain.TaskType]executor.TaskExecutor)

	return &migrator{
		jobService:    jobService,
		jobExecutors:  jobExecutors,
		taskExecutors: taskExecutors,
		slackClient:   slackClient,
		pollInterval:  cfg.MigratorPollInterval,
		cfg:           cfg,
		appClients:    appClients,
		// Semaphore to limit concurrent migrations
		semaphore: make(chan struct{}, cfg.MigratorMaxConcurrentExecutions),
	}
}

// SetTopicCache sets the topic cache and initializes task executors
func (mig *migrator) SetTopicCache(topicCache *cache.TopicCache) {
	mig.topicCache = topicCache
	mig.taskExecutors = getTaskExecutors(mig.jobService, mig.appClients, mig.cfg, topicCache)
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
