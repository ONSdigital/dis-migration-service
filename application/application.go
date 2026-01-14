package application

import (
	"context"
	"fmt"
	"time"

	sort "github.com/ONSdigital/dis-migration-service/api/sort"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dis-migration-service/statemachine"
	"github.com/ONSdigital/dis-migration-service/store"
	"github.com/ONSdigital/log.go/v2/log"
)

// JobService defines the contract for job-related operations
//
//go:generate moq -out mock/jobservice.go -pkg mock . JobService
type JobService interface {
	CreateJob(ctx context.Context, jobConfig *domain.JobConfig, userID string) (*domain.Job, error)
	GetJob(ctx context.Context, jobNumber int) (*domain.Job, error)
	ClaimJob(ctx context.Context) (*domain.Job, error)
	UpdateJobState(ctx context.Context, jobNumber int, newState domain.State, userID string) error
	GetJobs(ctx context.Context, field sort.SortParameterField, direction sort.SortParameterDirection, states []domain.State, limit, offset int) ([]*domain.Job, int, error)
	GetJobTasks(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error)
	CreateTask(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error)
	UpdateTask(ctx context.Context, task *domain.Task) error
	UpdateTaskState(ctx context.Context, taskID string, newState domain.State) error
	ClaimTask(ctx context.Context) (*domain.Task, error)
	CountTasksByJobNumber(ctx context.Context, jobNumber int) (int, error)
	GetNextJobNumber(ctx context.Context) (*domain.Counter, error)
	CreateEvent(ctx context.Context, jobNumber int, event *domain.Event) (*domain.Event, error)
	GetJobEvents(ctx context.Context, jobNumber int, limit, offset int) ([]*domain.Event, int, error)
	CountEventsByJobNumber(ctx context.Context, jobNumber int) (int, error)
}

type jobService struct {
	store   *store.Datastore
	clients *clients.ClientList
	config  *config.Config
}

// Setup initializes a new JobService with the provided
// dependencies.
func Setup(datastore *store.Datastore, appClients *clients.ClientList, cfg *config.Config) JobService {
	return &jobService{
		store:   datastore,
		clients: appClients,
		config:  cfg,
	}
}

// CreateJob creates a new migration job based on the
// provided job configuration and logs an event with the requesting user's ID.
func (js *jobService) CreateJob(ctx context.Context, jobConfig *domain.JobConfig, userID string) (*domain.Job, error) {
	label, err := jobConfig.ValidateExternal(ctx, *js.clients)
	if err != nil {
		return &domain.Job{}, err
	}

	// increment and get the job number counter
	jobNumberCounter, err := js.GetNextJobNumber(ctx)
	if err != nil {
		log.Error(ctx, "failed to get next job number counter", err)
		return &domain.Job{}, appErrors.ErrInternalServerError
	}

	// Create job with label
	job := domain.NewJob(jobConfig, jobNumberCounter.CounterValue, label)

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

	// Log job submission event
	if js.config.EnableEventLogging {
		if err := js.logJobEvent(ctx, job.JobNumber, string(domain.StateSubmitted), userID); err != nil {
			log.Error(ctx, "failed to log job submission event", err, log.Data{
				"job_number": job.JobNumber,
			})
			// TODO: Consider implementing a notification mechanism (e.g., alerts) for event logging failure
		}
	}

	return &job, nil
}

// GetNextJobNumber increments the job number counter,
// in mongoDB, and then returns it.
func (js *jobService) GetNextJobNumber(ctx context.Context) (*domain.Counter, error) {
	return js.store.GetNextJobNumberCounter(ctx)
}

// GetJob retrieves a migration job by its job number.
func (js *jobService) GetJob(ctx context.Context, jobNumber int) (*domain.Job, error) {
	return js.store.GetJob(ctx, jobNumber)
}

// UpdateJobState updates the state of a migration job and logs
// an event with the requesting user's ID.
func (js *jobService) UpdateJobState(ctx context.Context, jobNumber int, newState domain.State, userID string) error {
	job, err := js.store.GetJob(ctx, jobNumber)
	if err != nil {
		return err
	}

	if err := statemachine.ValidateTransition(job.State, newState); err != nil {
		return err
	}

	now := time.Now().UTC()
	err = js.store.UpdateJobState(ctx, job.ID, newState, now)
	if err != nil {
		return fmt.Errorf("failed to update job state: %w", err)
	}

	// Log event for approval or rejected state transitions if feature is enabled
	if js.config.EnableEventLogging && (newState == domain.StateApproved || newState == domain.StateRejected) {
		if err := js.logJobEvent(ctx, jobNumber, string(newState), userID); err != nil {
			log.Error(ctx, "failed to log job state transition event", err, log.Data{
				"job_number": jobNumber,
				"new_state":  newState,
			})
			// TODO: Consider implementing a notification mechanism (e.g., alerts) for event logging failure
		}
	}
	return nil
}

// GetJobs retrieves a list of migration jobs with pagination.
func (js *jobService) GetJobs(ctx context.Context, field sort.SortParameterField, direction sort.SortParameterDirection, states []domain.State, limit, offset int) ([]*domain.Job, int, error) {
	return js.store.GetJobs(ctx, field, direction, states, limit, offset)
}

// ClaimJob claims a pending job for processing.
func (js *jobService) ClaimJob(ctx context.Context) (*domain.Job, error) {
	transitions := []struct {
		from domain.State
		to   domain.State
	}{
		{from: domain.StateSubmitted, to: domain.StateMigrating},
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
func (js *jobService) UpdateTaskState(ctx context.Context, taskID string, newState domain.State) error {
	task, err := js.store.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	// Validate state transition
	if err := statemachine.ValidateTransition(task.State, newState); err != nil {
		return err
	}

	// Update in database
	now := time.Now().UTC()
	err = js.store.UpdateTaskState(ctx, taskID, newState, now)
	if err != nil {
		return fmt.Errorf("failed to update task state: %w", err)
	}

	return nil
}

// ClaimTask claims a pending task for processing.
func (js *jobService) ClaimTask(ctx context.Context) (*domain.Task, error) {
	transitions := []struct {
		from domain.State
		to   domain.State
	}{
		{from: domain.StateSubmitted, to: domain.StateMigrating},
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
func (js *jobService) GetJobTasks(ctx context.Context, states []domain.State, jobNumber, limit, offset int) ([]*domain.Task, int, error) {
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

// logJobEvent creates and stores an event for a job state transition.
func (js *jobService) logJobEvent(ctx context.Context, jobNumber int, action, userID string) error {
	// Create the event using domain factory
	event := domain.NewEvent(jobNumber, action, userID)

	// Store the event
	if err := js.store.CreateEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	log.Info(ctx, "job event logged successfully", log.Data{
		"job_number": jobNumber,
		"action":     action,
		"user_id":    event.RequestedBy.ID,
		"event_id":   event.ID,
	})

	return nil
}
