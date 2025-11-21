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

// CreateJob creates a new migration job.
func (m *Mongo) CreateJob(ctx context.Context, job *domain.Job) error {
	_, err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollectionTitle)).InsertOne(ctx, job)
	if err != nil {
		return err
	}

	return nil
}

// GetJob retrieves a job by its ID.
func (m *Mongo) GetJob(ctx context.Context, jobID string) (*domain.Job, error) {
	var job domain.Job
	if err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollectionTitle)).
		FindOne(ctx, bson.M{"_id": jobID}, &job); err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) {
			return nil, appErrors.ErrJobNotFound
		}
		return nil, appErrors.ErrInternalServerError
	}
	return &job, nil
}

// GetJobs retrieves a list of migration jobs with pagination.
func (m *Mongo) GetJobs(ctx context.Context, limit, offset int) ([]*domain.Job, int, error) {
	var results []*domain.Job

	totalCount, err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollectionTitle)).
		Find(
			ctx, bson.M{},
			&results,
			mongodriver.Limit(limit), mongodriver.Offset(offset),
			mongodriver.Sort(bson.M{"last_updated": -1}),
		)
	return results, totalCount, err
}

// GetJobsByConfigAndState retrieves jobs based on the provided job
// configuration and states.
func (m *Mongo) GetJobsByConfigAndState(ctx context.Context, jc *domain.JobConfig, stateFilter []domain.JobState, limit, offset int) ([]*domain.Job, error) {
	var results []*domain.Job

	_, err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollectionTitle)).
		Find(
			ctx,
			bson.M{
				"config": &jc,
				"state":  bson.M{"$in": stateFilter},
			},
			&results,
			mongodriver.Limit(limit), mongodriver.Offset(offset),
		)

	return results, err
}

// GetJobsByConfig retrieves jobs based on the provided job configuration.
func (m *Mongo) GetJobsByConfig(ctx context.Context, jc *domain.JobConfig, limit, offset int) ([]*domain.Job, error) {
	return m.GetJobsByConfigAndState(ctx, jc, []domain.JobState{}, limit, offset)
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
