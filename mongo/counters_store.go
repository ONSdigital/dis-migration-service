package mongo

import (
	"context"
	"errors"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

// GetJobNumberCounter retrieves the current value from the JobNumberCounter.
func (m *Mongo) GetJobNumberCounter(ctx context.Context) (*domain.Counters, error) {
	var jobNumberCounter domain.Counters
	if err := m.Connection.Collection(m.ActualCollectionName(config.CountersCollectionTitle)).
		FindOne(ctx, bson.M{"counter_name": "job_number_counter"}, &jobNumberCounter); err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return nil, appErrors.ErrJobNumberCounterNotFound
		}
		return nil, appErrors.ErrInternalServerError
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
