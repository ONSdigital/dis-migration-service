package store

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

// Datastore provides a datastore.Storer interface used to store, retrieve, remove or update bundles
//
//go:generate moq -out mock/mongo.go -pkg mock . MongoDB
//go:generate moq -out mock/datastore.go -pkg mock . Storer

type Datastore struct {
	Backend Storer
}

type dataMongoDB interface {

	// Jobs

	CreateJob(ctx context.Context, job *domain.Job) (*domain.Job, error)
	GetJob(ctx context.Context, jobID string) (*domain.Job, error)

	// TODO: Tasks
	// TODO: Events
	CreateEvent(ctx context.Context, event *domain.Event) error

	// Other
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Close(ctx context.Context) error
}

// MongoDB represents all the required methods from mongo DB
type MongoDB interface {
	dataMongoDB
	Close(context.Context) error
	Checker(context.Context, *healthcheck.CheckState) error
}

// Storer represents basic data access via Get, Remove and Upsert methods, abstracting it from mongoDB or graphDB
type Storer interface {
	dataMongoDB
}

func (ds *Datastore) CreateJob(ctx context.Context, job *domain.Job) (*domain.Job, error) {
	return ds.Backend.CreateJob(ctx, job)
}

func (ds *Datastore) GetJob(ctx context.Context, jobID string) (*domain.Job, error) {
	return ds.Backend.GetJob(ctx, jobID)
}
