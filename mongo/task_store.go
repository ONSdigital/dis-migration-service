package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateTask creates a new migration task.
func (m *Mongo) CreateTask(ctx context.Context, task *domain.Task) error {
	_, err := m.Connection.Collection(m.ActualCollectionName(config.TasksCollectionTitle)).InsertOne(ctx, task)
	if err != nil {
		return appErrors.ErrInternalServerError
	}

	return nil
}

// GetTask retrieves a task by its ID.
func (m *Mongo) GetTask(ctx context.Context, taskID string) (*domain.Task, error) {
	var task domain.Task
	if err := m.Connection.Collection(m.ActualCollectionName(config.TasksCollectionTitle)).
		FindOne(ctx, bson.M{"_id": taskID}, &task); err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) || errors.Is(err, mongo.ErrNoDocuments) {
			return nil, appErrors.ErrTaskNotFound
		}
		return nil, appErrors.ErrInternalServerError
	}
	return &task, nil
}

// UpdateTask updates an existing migration task.
func (m *Mongo) UpdateTask(ctx context.Context, task *domain.Task) error {
	filter := bson.M{"_id": task.ID}
	update := bson.M{"$set": task}

	result, err := m.Connection.Collection(m.ActualCollectionName(config.TasksCollectionTitle)).UpdateOne(ctx, filter, update)
	if err != nil {
		return appErrors.ErrInternalServerError
	}

	if result.MatchedCount == 0 {
		return appErrors.ErrTaskNotFound
	}

	return nil
}

// UpdateTaskState updates the state of a task and returns the updated task.
func (m *Mongo) UpdateTaskState(ctx context.Context, taskID string, newState domain.State, lastUpdated time.Time) error {
	collectionName := m.ActualCollectionName(config.TasksCollectionTitle)

	filter := bson.M{
		"_id": taskID,
	}
	update := bson.M{
		"$set": bson.M{
			"state":        newState,
			"last_updated": lastUpdated,
		},
	}

	// Update the document
	result, err := m.Connection.Collection(collectionName).UpdateOne(
		ctx,
		filter,
		update,
	)
	if err != nil {
		return appErrors.ErrInternalServerError
	}

	// Check if a document was found and updated
	if result.MatchedCount == 0 {
		return appErrors.ErrTaskNotFound
	}

	return nil
}

// ClaimTask claims a pending task for processing.
func (m *Mongo) ClaimTask(ctx context.Context, pendingState, activeState domain.State) (*domain.Task, error) {
	var task domain.Task

	filter := bson.M{"state": pendingState}
	update := bson.M{
		"$set": bson.M{
			"state":        activeState,
			"last_updated": time.Now(),
		},
	}

	err := m.Connection.Collection(m.ActualCollectionName(config.TasksCollectionTitle)).
		FindOneAndUpdate(ctx, filter, update, &task, mongodriver.ReturnDocument(options.After))
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) || errors.Is(err, mongo.ErrNoDocuments) {
			// If no pending jobs, no error.
			return nil, nil
		}
		return nil, err
	}

	return &task, nil
}

// GetJobTasks retrieves a list of migration tasks for a job with pagination.
func (m *Mongo) GetJobTasks(ctx context.Context, stateFilter []domain.State, jobNumber, limit, offset int) ([]*domain.Task, int, error) {
	var results []*domain.Task

	filter := bson.M{"job_number": jobNumber}

	if len(stateFilter) > 0 {
		filter["state"] = bson.M{"$in": stateFilter}
	}

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

// CountTasksByJobNumber returns the total count of tasks for a job.
func (m *Mongo) CountTasksByJobNumber(ctx context.Context, jobNumber int) (int, error) {
	filter := bson.M{"job_number": jobNumber}

	// Use Find to get the total count without actually retrieving documents
	var results []*domain.Task
	totalCount, err := m.Connection.Collection(m.ActualCollectionName(config.TasksCollectionTitle)).
		Find(ctx, filter, &results, mongodriver.Limit(1))

	if err != nil {
		return 0, appErrors.ErrInternalServerError
	}

	return totalCount, nil
}
