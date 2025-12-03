package application

import (
	"context"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dis-migration-service/store"
	"github.com/ONSdigital/log.go/v2/log"
)

// JobService defines the contract for job-related operations
//
//go:generate moq -out mock/jobservice.go -pkg mock . JobService
type JobService interface {
	CreateJob(ctx context.Context, jobConfig *domain.JobConfig) (*domain.Job, error)
	GetJob(ctx context.Context, jobID string) (*domain.Job, error)
	GetJobs(ctx context.Context, limit, offset int) ([]*domain.Job, int, error)
	GetJobTasks(ctx context.Context, jobID string, limit, offset int) ([]*domain.Task, int, error)
	CountTasksByJobID(ctx context.Context, jobID string) (int, error)
	CreateJobNumberCounter(ctx context.Context) error
}

type jobService struct {
	store   *store.Datastore
	clients *clients.ClientList
}

// Setup initializes a new JobService with the provided
// dependencies.
func Setup(datastore *store.Datastore, appClients *clients.ClientList) JobService {
	return &jobService{
		store:   datastore,
		clients: appClients,
	}
}

// CreateJob creates a new migration job based on the
// provided job configuration.
func (js *jobService) CreateJob(ctx context.Context, jobConfig *domain.JobConfig) (*domain.Job, error) {
	err := jobConfig.ValidateExternal(ctx, *js.clients)
	if err != nil {
		return &domain.Job{}, err
	}

	job := domain.NewJob(jobConfig)

	foundJobs, err := js.store.GetJobsByConfigAndState(ctx, job.Config, domain.GetNonCancelledStates(), 1, 0)
	if err != nil {
		log.Error(ctx, "failed to validate job creation", err)
		return &domain.Job{}, appErrors.ErrInternalServerError
	}
	if len(foundJobs) > 0 {
		log.Error(ctx, "found running jobs with this config", err, log.Data{
			"job_config": job.Config,
		})
		return &domain.Job{}, appErrors.ErrJobAlreadyRunning
	}

	err = js.store.CreateJob(ctx, &job)
	if err != nil {
		log.Error(ctx, "failed to create job", err)
		return &domain.Job{}, appErrors.ErrInternalServerError
	}
	return &job, nil
}

// CreateJobNumberCounter will create a counter with the following values:
// counter_name = "job_number_counter"
// counter_value = "0"
func (js *jobService) CreateJobNumberCounter(ctx context.Context) error {
	return js.store.CreateJobNumberCounter(ctx)
}

// GetJob retrieves a migration job by its ID.
func (js *jobService) GetJob(ctx context.Context, jobID string) (*domain.Job, error) {
	return js.store.GetJob(ctx, jobID)
}

// GetJobs retrieves a list of migration jobs with pagination.
func (js *jobService) GetJobs(ctx context.Context, limit, offset int) ([]*domain.Job, int, error) {
	return js.store.GetJobs(ctx, limit, offset)
}

// GetJobTasks retrieves a list of migration tasks for a job with pagination.
func (js *jobService) GetJobTasks(ctx context.Context, jobID string, limit, offset int) ([]*domain.Task, int, error) {
	return js.store.GetJobTasks(ctx, jobID, limit, offset)
}

// CountTasksByJobID returns the total count of tasks for a job.
func (js *jobService) CountTasksByJobID(ctx context.Context, jobID string) (int, error) {
	return js.store.CountTasksByJobID(ctx, jobID)
}
