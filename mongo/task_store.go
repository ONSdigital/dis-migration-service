package mongo

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

// CreateTask creates a new migration task.
func (m *Mongo) CreateTask(ctx context.Context, task *domain.Task) error {
	_, err := m.Connection.Collection(m.ActualCollectionName(config.TasksCollectionTitle)).InsertOne(ctx, task)
	if err != nil {
		return appErrors.ErrInternalServerError
	}

	return nil
}

// GetJobTasks retrieves a list of migration tasks for a job with pagination.
func (m *Mongo) GetJobTasks(ctx context.Context, jobID string, limit, offset int) ([]*domain.Task, int, error) {
	var results []*domain.Task

	filter := bson.M{"job_id": jobID}

	totalCount, err := m.Connection.Collection(m.ActualCollectionName(config.TasksCollectionTitle)).
		Find(
			ctx,
			filter,
			&results,
			mongodriver.Limit(limit),
			mongodriver.Offset(offset),
			mongodriver.Sort(bson.M{"last_updated": -1}),
		)

	if err != nil {
		return nil, 0, appErrors.ErrInternalServerError
	}

	return results, totalCount, nil
}

// CountTasksByJobID returns the total count of tasks for a job.
func (m *Mongo) CountTasksByJobID(ctx context.Context, jobID string) (int, error) {
	filter := bson.M{"job_id": jobID}

	// Use Find to get the total count without actually retrieving documents
	var results []*domain.Task
	totalCount, err := m.Connection.Collection(m.ActualCollectionName(config.TasksCollectionTitle)).
		Find(ctx, filter, &results, mongodriver.Limit(1))

	if err != nil {
		return 0, appErrors.ErrInternalServerError
	}

	return totalCount, nil
}
