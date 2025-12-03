package mongo

import (
	"context"
	"errors"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
)

// CreateJobNumberCounter creates a new Counter to use for creating unique human-readable job numbers.
// NB. This should only be used if the JobNumberCounter does not already exist.
func (m *Mongo) CreateJobNumberCounter(ctx context.Context) error {
	// Firstly check that the Job Number Counter does not already exist
	counterRetrieved, err := m.GetJobNumberCounter(ctx)
	if counterRetrieved != nil {
		logData := log.Data{"jobNumberCounter": counterRetrieved}
		log.Info(ctx, "job number counter already exists in mongo DB", logData)
		return appErrors.ErrJobNumberCounterAlreadyExists
	}

	jobNumberCounter := &domain.Counter{
		CounterName:  "job_number_counter",
		CounterValue: "0",
	}
	logData := log.Data{"jobNumberCounter": jobNumberCounter}
	log.Info(ctx, "creating job number counter in mongo DB", logData)

	_, err = m.Connection.Collection(m.ActualCollectionName(config.CountersCollectionTitle)).InsertOne(ctx, jobNumberCounter)
	if err != nil {
		log.Error(ctx, "failed to insert job number counter into mongo DB", err, logData)
		return err
	}

	return nil
}

// GetJobNumberCounter retrieves the current value from the JobNumberCounter.
// If the JobNumberCounter does not exist then it creates it.
func (m *Mongo) GetJobNumberCounter(ctx context.Context) (*domain.Counter, error) {
	var jobNumberCounter domain.Counter
	if err := m.Connection.Collection(m.ActualCollectionName(config.CountersCollectionTitle)).
		FindOne(ctx, bson.M{"counter_name": "job_number_counter"}, &jobNumberCounter); err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			log.Error(ctx, "the job number counter does not exist so shall create it",
				appErrors.ErrJobNumberCounterNotFound)
			err = m.CreateJobNumberCounter(ctx)
			if err != nil {
				log.Error(ctx, "error creating job number counter", err)
				return nil, err
			}
		} else {
			return nil, appErrors.ErrInternalServerError
		}
	}
	return &jobNumberCounter, nil
}

// UpdateJobNumberCounter updates the job number counter with the number provided in the updates
func (m *Mongo) UpdateJobNumberCounter(ctx context.Context, updates bson.M) error {
	update := bson.M{"$set": updates}

	if _, err := m.Connection.Collection(m.ActualCollectionName(config.CountersCollectionTitle)).Must().
		UpdateOne(ctx, bson.M{"counter_name": "job_number_counter"}, update); err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return appErrors.ErrJobNumberCounterNotFound
		}
		return appErrors.ErrInternalServerError
	}
	return nil
}

//// GetJobNumberCounterAndUpdate finds and updates the job number counter with the number provided in the updates
//func (m *Mongo) GetJobNumberCounterAndUpdate(ctx context.Context, updates bson.M) error {
//}
