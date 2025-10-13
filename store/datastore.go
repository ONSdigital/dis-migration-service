package store

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

// Datastore provides a datastore.Storer interface used to store, retrieve, remove or update migrations
//
//go:generate moq -out datastoretest/mongo.go -pkg storetest . MongoDB
//go:generate moq -out datastoretest/datastore.go -pkg storetest . Storer

type Datastore struct {
	Backend Storer
}

// MongoDB represents all the required methods from mongo DB
type MongoDB interface {
	Close(context.Context) error
	Checker(context.Context, *healthcheck.CheckState) error
}

// Storer represents basic data access via Get, Remove and Upsert methods, abstracting it from mongoDB
type Storer interface {
	MongoDB
}
