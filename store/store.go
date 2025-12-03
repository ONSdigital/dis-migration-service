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
	GetJob(ctx context.Context, jobID string) (*domain.Job, error)
	GetJobs(ctx context.Context, limit, offset int) ([]*domain.Job, int, error)
	GetJobsByConfigAndState(ctx context.Context, jc *domain.JobConfig, states []domain.JobState, limit, offset int) ([]*domain.Job, error)
	CreateJobNumberCounter(ctx context.Context) error

	// Tasks
	GetJobTasks(ctx context.Context, jobID string, limit, offset int) ([]*domain.Task, int, error)
	CountTasksByJobID(ctx context.Context, jobID string) (int, error)

	// Events
	CreateEvent(ctx context.Context, event *domain.Event) error

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

// GetJob retrieves a job by its ID.
func (ds *Datastore) GetJob(ctx context.Context, jobID string) (*domain.Job, error) {
	return ds.Backend.GetJob(ctx, jobID)
}

// GetJobs retrieves a list of migration jobs with pagination.
func (ds *Datastore) GetJobs(ctx context.Context, limit, offset int) ([]*domain.Job, int, error) {
	return ds.Backend.GetJobs(ctx, limit, offset)
}

// GetJobsByConfigAndState retrieves jobs based on the provided job
// configuration and states.
func (ds *Datastore) GetJobsByConfigAndState(ctx context.Context, jc *domain.JobConfig, states []domain.JobState, limit, offset int) ([]*domain.Job, error) {
	return ds.Backend.GetJobsByConfigAndState(ctx, jc, states, limit, offset)
}

// GetJobTasks retrieves a list of migration tasks for a job with pagination.
func (ds *Datastore) GetJobTasks(ctx context.Context, jobID string, limit, offset int) ([]*domain.Task, int, error) {
	return ds.Backend.GetJobTasks(ctx, jobID, limit, offset)
}

// CountTasksByJobID returns the total count of tasks for a job.
func (ds *Datastore) CountTasksByJobID(ctx context.Context, jobID string) (int, error) {
	return ds.Backend.CountTasksByJobID(ctx, jobID)
}

func (ds *Datastore) CreateJobNumberCounter(ctx context.Context) error {
	return ds.Backend.CreateJobNumberCounter(ctx)
}
