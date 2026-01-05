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
	CreateJob(ctx context.Context, jobConfig *domain.JobConfig, jobNumberCounterValue int) (*domain.Job, error)
	GetJob(ctx context.Context, jobNumber int) (*domain.Job, error)
	ClaimJob(ctx context.Context) (*domain.Job, error)
	UpdateJobState(ctx context.Context, jobNumber int, newState domain.JobState) error
	GetJobs(ctx context.Context, states []domain.JobState, limit, offset int) ([]*domain.Job, int, error)
	GetJobTasks(ctx context.Context, states []domain.TaskState, jobNumber int, limit, offset int) ([]*domain.Task, int, error)
	CreateTask(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error)
	UpdateTask(ctx context.Context, task *domain.Task) error
	UpdateTaskState(ctx context.Context, taskID string, newState domain.TaskState) error
	ClaimTask(ctx context.Context) (*domain.Task, error)
	CountTasksByJobNumber(ctx context.Context, jobNumber int) (int, error)
	GetJobNumberCounter(ctx context.Context) (*domain.Counter, error)
	GetNextJobNumberCounter(ctx context.Context) (*domain.Counter, error)
	CreateEvent(ctx context.Context, jobNumber int, event *domain.Event) (*domain.Event, error)
	GetJobEvents(ctx context.Context, jobNumber int, limit, offset int) ([]*domain.Event, int, error)
	CountEventsByJobNumber(ctx context.Context, jobNumber int) (int, error)
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
func (js *jobService) CreateJob(ctx context.Context, jobConfig *domain.JobConfig, jobNumberCounterValue int) (*domain.Job, error) {
	label, err := jobConfig.ValidateExternal(ctx, *js.clients)
	if err != nil {
		return &domain.Job{}, err
	}

	// Create job with label
	job := domain.NewJob(jobConfig, jobNumberCounterValue, label)

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

// GetJobNumberCounter will return the domain.Counter with
// counter_name = "job_number_counter"
func (js *jobService) GetJobNumberCounter(ctx context.Context) (*domain.Counter, error) {
	return js.store.GetJobNumberCounter(ctx)
}

// GetNextJobNumberCounter increments the job number counter,
// in mongoDB, and then returns it
func (js *jobService) GetNextJobNumberCounter(ctx context.Context) (*domain.Counter, error) {
	return js.store.GetNextJobNumberCounter(ctx)
}

// GetJob retrieves a migration job by its job number.
func (js *jobService) GetJob(ctx context.Context, jobNumber int) (*domain.Job, error) {
	return js.store.GetJob(ctx, jobNumber)
}

// UpdateJobState updates the state of a migration job.
func (js *jobService) UpdateJobState(ctx context.Context, jobNumber int, newState domain.JobState) error {
	//TODO: validate state transition
	job, err := js.store.GetJob(ctx, jobNumber)
	if err != nil {
		return err
	}
	job.State = newState

	err = js.store.UpdateJob(ctx, job)
	if err != nil {
		log.Error(ctx, "failed to update job", err, log.Data{
			"job_id":    job.ID,
			"new_state": newState,
		})
		return appErrors.ErrInternalServerError
	}

	return nil
}

// GetJobs retrieves a list of migration jobs with pagination.
func (js *jobService) GetJobs(ctx context.Context, states []domain.JobState, limit, offset int) ([]*domain.Job, int, error) {
	return js.store.GetJobs(ctx, states, limit, offset)
}

// ClaimJob claims a pending job for processing.
func (js *jobService) ClaimJob(ctx context.Context) (*domain.Job, error) {
	transitions := []struct {
		from domain.JobState
		to   domain.JobState
	}{
		{from: domain.JobStateSubmitted, to: domain.JobStateMigrating},
	}

	for _, tr := range transitions {
		job, err := js.store.ClaimJob(ctx, tr.from, tr.to)
		if err != nil {
			return nil, err
		}
		if job != nil {
			return job, nil
		}
	}
	return nil, nil
}

// CreateTask creates a new migration task for a job.
func (js *jobService) CreateTask(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
	// Verify job exists
	_, err := js.store.GetJob(ctx, jobNumber)
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

// UpdateTask updates a migration task.
func (js *jobService) UpdateTask(ctx context.Context, task *domain.Task) error {
	err := js.store.UpdateTask(ctx, task)
	if err != nil {
		log.Error(ctx, "failed to update task", err, log.Data{
			"task_id": task.ID,
		})
		return appErrors.ErrInternalServerError
	}

	return nil
}

// UpdateTaskState updates the state of a migration task.
func (js *jobService) UpdateTaskState(ctx context.Context, taskID string, newState domain.TaskState) error {
	//TODO: validate state transition
	task, err := js.store.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	task.State = newState
	err = js.store.UpdateTask(ctx, task)
	if err != nil {
		log.Error(ctx, "failed to update task", err, log.Data{
			"task_id":   task.ID,
			"new_state": newState,
		})
		return appErrors.ErrInternalServerError
	}

	return nil
}

// ClaimTask claims a pending task for processing.
func (js *jobService) ClaimTask(ctx context.Context) (*domain.Task, error) {
	transitions := []struct {
		from domain.TaskState
		to   domain.TaskState
	}{
		{from: domain.TaskStateSubmitted, to: domain.TaskStateMigrating},
	}

	for _, tr := range transitions {
		task, err := js.store.ClaimTask(ctx, tr.from, tr.to)
		if err != nil {
			return nil, err
		}
		if task != nil {
			return task, nil
		}
	}
	return nil, nil
}

// GetJobTasks retrieves a list of migration tasks for a job with pagination.
func (js *jobService) GetJobTasks(ctx context.Context, states []domain.TaskState, jobNumber, limit, offset int) ([]*domain.Task, int, error) {
	return js.store.GetJobTasks(ctx, states, jobNumber, limit, offset)
}

// CountTasksByJobNumber returns the total count of tasks for a job.
func (js *jobService) CountTasksByJobNumber(ctx context.Context, jobNumber int) (int, error) {
	return js.store.CountTasksByJobNumber(ctx, jobNumber)
}

// CreateEvent creates a new migration event for a job.
func (js *jobService) CreateEvent(ctx context.Context, jobNumber int, event *domain.Event) (*domain.Event, error) {
	// Verify job exists
	_, err := js.store.GetJob(ctx, jobNumber)
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
func (js *jobService) GetJobEvents(ctx context.Context, jobNumber, limit, offset int) ([]*domain.Event, int, error) {
	return js.store.GetJobEvents(ctx, jobNumber, limit, offset)
}

// CountEventsByJobNumber returns the total count of events for a job.
func (js *jobService) CountEventsByJobNumber(ctx context.Context, jobNumber int) (int, error) {
	return js.store.CountEventsByJobNumber(ctx, jobNumber)
}
