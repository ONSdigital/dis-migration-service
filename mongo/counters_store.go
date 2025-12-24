package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
)

/* createJobNumberCounter creates a new Counter to use for creating unique
* human-readable job numbers.
* The new Counter will contain the following values:
* counter_name = "job_number_counter"
* counter_value = "0"
*
* NB. This private function should only be called by GetJobNumberCounter,
* which will only call it if a JobNumberCounter does not already exist. */
func (m *Mongo) createJobNumberCounter(ctx context.Context) error {
	jobNumberCounter := &domain.Counter{
		CounterName:  "job_number_counter",
		CounterValue: 0,
	}
	logData := log.Data{"jobNumberCounter": jobNumberCounter}
	log.Info(ctx, "creating job number counter in mongo DB", logData)

	_, err := m.Connection.Collection(m.ActualCollectionName(config.CountersCollectionTitle)).InsertOne(ctx, jobNumberCounter)
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
			err = m.createJobNumberCounter(ctx)
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

// UpdateJobNumberCounter increments the job number counter, in mongoDB, by 1
func (m *Mongo) UpdateJobNumberCounter(ctx context.Context) error {
	_, err := m.Connection.Collection(m.ActualCollectionName(config.CountersCollectionTitle)).UpdateOne(ctx, bson.M{
		"counter_name": "job_number_counter",
	}, bson.D{
		{Key: "$inc", Value: bson.D{primitive.E{Key: "counter_value", Value: 1}}},
	})

	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return appErrors.ErrJobNumberCounterNotFound
		}
		return appErrors.ErrInternalServerError
	}

	return nil
}
