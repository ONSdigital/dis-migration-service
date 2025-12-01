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
	GetJobs(ctx context.Context, states []domain.JobState, limit, offset int) ([]*domain.Job, int, error)
	CreateTask(ctx context.Context, jobID string, task *domain.Task) (*domain.Task, error)
	GetJobTasks(ctx context.Context, jobID string, limit, offset int) ([]*domain.Task, int, error)
	CountTasksByJobID(ctx context.Context, jobID string) (int, error)
	CreateEvent(ctx context.Context, jobID string, event *domain.Event) (*domain.Event, error)
	GetJobEvents(ctx context.Context, jobID string, limit, offset int) ([]*domain.Event, int, error)
	CountEventsByJobID(ctx context.Context, jobID string) (int, error)
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

// GetJob retrieves a migration job by its ID.
func (js *jobService) GetJob(ctx context.Context, jobID string) (*domain.Job, error) {
	return js.store.GetJob(ctx, jobID)
}

// GetJobs retrieves a list of migration jobs with pagination.
func (js *jobService) GetJobs(ctx context.Context, states []domain.JobState, limit, offset int) ([]*domain.Job, int, error) {
	return js.store.GetJobs(ctx, states, limit, offset)
}

// CreateTask creates a new migration task for a job.
func (js *jobService) CreateTask(ctx context.Context, jobID string, task *domain.Task) (*domain.Task, error) {
	// Verify job exists
	_, err := js.store.GetJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	// TODO: Validate job is in a state where tasks can be created

	// Create the task in the store
	err = js.store.CreateTask(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetJobTasks retrieves a list of migration tasks for a job with pagination.
func (js *jobService) GetJobTasks(ctx context.Context, jobID string, limit, offset int) ([]*domain.Task, int, error) {
	return js.store.GetJobTasks(ctx, jobID, limit, offset)
}

// CountTasksByJobID returns the total count of tasks for a job.
func (js *jobService) CountTasksByJobID(ctx context.Context, jobID string) (int, error) {
	return js.store.CountTasksByJobID(ctx, jobID)
}

// CreateEvent creates a new migration event for a job.
func (js *jobService) CreateEvent(ctx context.Context, jobID string, event *domain.Event) (*domain.Event, error) {
	// Verify job exists
	_, err := js.store.GetJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	// Create the event in the store
	err = js.store.CreateEvent(ctx, event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// GetJobEvents retrieves a list of migration events for a job with pagination.
func (js *jobService) GetJobEvents(ctx context.Context, jobID string, limit, offset int) ([]*domain.Event, int, error) {
	return js.store.GetJobEvents(ctx, jobID, limit, offset)
}

// CountEventsByJobID returns the total count of events for a job.
func (js *jobService) CountEventsByJobID(ctx context.Context, jobID string) (int, error) {
	return js.store.CountEventsByJobID(ctx, jobID)
}
