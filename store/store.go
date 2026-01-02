package store

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/mongo.go -pkg mock . MongoDB
//go:generate moq -out mock/datastore.go -pkg mock . Storer

// Datastore provides a datastore.Storer interface used to store,
// retrieve, remove or update bundles
type Datastore struct {
	Backend Storer
}

type dataMongoDB interface {
	// Jobs
	CreateJob(ctx context.Context, job *domain.Job) error
	GetJob(ctx context.Context, jobNumber int) (*domain.Job, error)
	GetJobs(ctx context.Context, states []domain.JobState, limit, offset int) ([]*domain.Job, int, error)
	ClaimJob(ctx context.Context, pendingState domain.JobState, activeState domain.JobState) (*domain.Job, error)
	GetJobsByConfigAndState(ctx context.Context, jc *domain.JobConfig, states []domain.JobState, limit, offset int) ([]*domain.Job, error)
	GetJobNumberCounter(ctx context.Context) (*domain.Counter, error)
	UpdateJobNumberCounter(ctx context.Context) error
	UpdateJob(ctx context.Context, job *domain.Job) error

	// Tasks
	CreateTask(ctx context.Context, task *domain.Task) error
	GetTask(ctx context.Context, taskID string) (*domain.Task, error)
	GetJobTasks(ctx context.Context, states []domain.TaskState, jobNumber int, limit, offset int) ([]*domain.Task, int, error)
	CountTasksByJobNumber(ctx context.Context, jobNumber int) (int, error)
	UpdateTask(ctx context.Context, task *domain.Task) error
	ClaimTask(ctx context.Context, pendingState domain.TaskState, activeState domain.TaskState) (*domain.Task, error)

	// Events
	CreateEvent(ctx context.Context, event *domain.Event) error
	GetJobEvents(ctx context.Context, jobNumber int, limit, offset int) ([]*domain.Event, int, error)
	CountEventsByJobNumber(ctx context.Context, jobNumber int) (int, error)

	// Other
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Close(ctx context.Context) error
}

// MongoDB represents all the required methods from mongoDB
type MongoDB interface {
	dataMongoDB
	Close(context.Context) error
	Checker(context.Context, *healthcheck.CheckState) error
}

// Storer represents basic data access via Get, Remove and Upsert methods,
// abstracting it from mongoDB or graphDB
type Storer interface {
	dataMongoDB
}

// CreateJob creates a new migration job.
func (ds *Datastore) CreateJob(ctx context.Context, job *domain.Job) error {
	return ds.Backend.CreateJob(ctx, job)
}

// GetJob retrieves a job by its job number.
func (ds *Datastore) GetJob(ctx context.Context, jobNumber int) (*domain.Job, error) {
	return ds.Backend.GetJob(ctx, jobNumber)
}

// ClaimJob claims a pending job for processing.
func (ds *Datastore) ClaimJob(ctx context.Context, pendingState, activeState domain.JobState) (*domain.Job, error) {
	return ds.Backend.ClaimJob(ctx, pendingState, activeState)
}

// UpdateJob updates an existing migration job.
func (ds *Datastore) UpdateJob(ctx context.Context, job *domain.Job) error {
	return ds.Backend.UpdateJob(ctx, job)
}

// GetJobs retrieves a list of migration jobs with pagination.
func (ds *Datastore) GetJobs(ctx context.Context, states []domain.JobState, limit, offset int) ([]*domain.Job, int, error) {
	return ds.Backend.GetJobs(ctx, states, limit, offset)
}

// GetJobsByConfigAndState retrieves jobs based on the provided job
// configuration and states.
func (ds *Datastore) GetJobsByConfigAndState(ctx context.Context, jc *domain.JobConfig, states []domain.JobState, limit, offset int) ([]*domain.Job, error) {
	return ds.Backend.GetJobsByConfigAndState(ctx, jc, states, limit, offset)
}

// CreateTask creates a new migration task.
func (ds *Datastore) CreateTask(ctx context.Context, task *domain.Task) error {
	return ds.Backend.CreateTask(ctx, task)
}

// GetTask retrieves a migration task by its ID.
func (ds *Datastore) GetTask(ctx context.Context, taskID string) (*domain.Task, error) {
	return ds.Backend.GetTask(ctx, taskID)
}

// UpdateTask updates an existing migration task.
func (ds *Datastore) UpdateTask(ctx context.Context, task *domain.Task) error {
	return ds.Backend.UpdateTask(ctx, task)
}

// ClaimTask claims a pending task for processing.
func (ds *Datastore) ClaimTask(ctx context.Context, pendingState, activeState domain.TaskState) (*domain.Task, error) {
	return ds.Backend.ClaimTask(ctx, pendingState, activeState)
}

// GetJobTasks retrieves a list of migration tasks for a job with pagination.
func (ds *Datastore) GetJobTasks(ctx context.Context, states []domain.TaskState, jobNumber, limit, offset int) ([]*domain.Task, int, error) {
	return ds.Backend.GetJobTasks(ctx, states, jobNumber, limit, offset)
}

// CountTasksByJobNumber returns the total count of tasks for a job.
func (ds *Datastore) CountTasksByJobNumber(ctx context.Context, jobNumber int) (int, error) {
	return ds.Backend.CountTasksByJobNumber(ctx, jobNumber)
}

// GetJobNumberCounter retrieves the current value from the JobNumberCounter.
// If the JobNumberCounter does not exist then it creates it.
func (ds *Datastore) GetJobNumberCounter(ctx context.Context) (*domain.Counter, error) {
	return ds.Backend.GetJobNumberCounter(ctx)
}

// UpdateJobNumberCounter increments the job number counter, in mongoDB, by 1.
func (ds *Datastore) UpdateJobNumberCounter(ctx context.Context) error {
	return ds.Backend.UpdateJobNumberCounter(ctx)
}

// CreateEvent creates a new migration event.
func (ds *Datastore) CreateEvent(ctx context.Context, event *domain.Event) error {
	return ds.Backend.CreateEvent(ctx, event)
}

// GetJobEvents retrieves a list of migration events for a job with pagination.
func (ds *Datastore) GetJobEvents(ctx context.Context, jobNumber, limit, offset int) ([]*domain.Event, int, error) {
	return ds.Backend.GetJobEvents(ctx, jobNumber, limit, offset)
}

// CountEventsByJobNumber returns the total count of events for a job.
func (ds *Datastore) CountEventsByJobNumber(ctx context.Context, jobNumber int) (int, error) {
	return ds.Backend.CountEventsByJobNumber(ctx, jobNumber)
}
