package mongo

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

// CreateEvent creates a new event in the database.
func (m *Mongo) CreateEvent(ctx context.Context, event *domain.Event) error {
	_, err := m.Connection.Collection(m.ActualCollectionName(config.EventsCollectionTitle)).InsertOne(ctx, event)
	if err != nil {
		return appErrors.ErrInternalServerError
	}

	return nil
}

// GetJobEvents retrieves a list of migration events for a job with pagination.
func (m *Mongo) GetJobEvents(ctx context.Context, jobNumber int, limit, offset int) ([]*domain.Event, int, error) {
	var results []*domain.Event

	filter := bson.M{"job_number": jobNumber}

	totalCount, err := m.Connection.Collection(m.ActualCollectionName(config.EventsCollectionTitle)).
		Find(
			ctx,
			filter,
			&results,
			mongodriver.Limit(limit),
			mongodriver.Offset(offset),
			mongodriver.Sort(bson.M{"created_at": -1}), // Sort by timestamp descending (newest first)
		)

	if err != nil {
		return nil, 0, appErrors.ErrInternalServerError
	}

	return results, totalCount, nil
}

// CountEventsByJobNumber returns the total count of events for a job.
func (m *Mongo) CountEventsByJobNumber(ctx context.Context, jobNumber int) (int, error) {
	filter := bson.M{"job_number": jobNumber}

	// Use Find to get the total count without actually retrieving documents
	var results []*domain.Event
	totalCount, err := m.Connection.Collection(m.ActualCollectionName(config.EventsCollectionTitle)).
		Find(ctx, filter, &results, mongodriver.Limit(1))

	if err != nil {
		return 0, appErrors.ErrInternalServerError
	}

	return totalCount, nil
}
