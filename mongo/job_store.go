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

// CreateJob creates a new migration job.
func (m *Mongo) CreateJob(ctx context.Context, job *domain.Job) error {
	_, err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollectionTitle)).InsertOne(ctx, job)
	if err != nil {
		return err
	}

	return nil
}

// GetJob retrieves a job by its job number.
func (m *Mongo) GetJob(ctx context.Context, jobNumber int) (*domain.Job, error) {
	var job domain.Job
	if err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollectionTitle)).
		FindOne(ctx, bson.M{"job_number": jobNumber}, &job); err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) || errors.Is(err, mongo.ErrNoDocuments) {
			return nil, appErrors.ErrJobNotFound
		}
		return nil, appErrors.ErrInternalServerError
	}
	return &job, nil
}

// GetJobs retrieves a list of migration jobs with pagination.
func (m *Mongo) GetJobs(ctx context.Context, stateFilter []domain.State, limit, offset int) ([]*domain.Job, int, error) {
	var results []*domain.Job

	filter := bson.M{}
	if len(stateFilter) > 0 {
		filter["state"] = bson.M{"$in": stateFilter}
	}

	totalCount, err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollectionTitle)).
		Find(
			ctx,
			filter,
			&results,
			mongodriver.Limit(limit),
			mongodriver.Offset(offset),
			mongodriver.Sort(bson.M{"last_updated": -1}),
		)

	return results, totalCount, err
}

// GetJobsByConfigAndState retrieves jobs based on the provided job
// configuration and states.
func (m *Mongo) GetJobsByConfigAndState(ctx context.Context, jc *domain.JobConfig, stateFilter []domain.State, limit, offset int) ([]*domain.Job, error) {
	var results []*domain.Job

	_, err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollectionTitle)).
		Find(
			ctx,
			bson.M{
				"config.source_id": jc.SourceID,
				"config.target_id": jc.TargetID,
				"config.type":      jc.Type,
				"state":            bson.M{"$in": stateFilter},
			},
			&results,
			mongodriver.Limit(limit), mongodriver.Offset(offset),
		)

	return results, err
}

// GetJobsByState retrieves a list of migration jobs filtered by their states.
func (m *Mongo) GetJobsByState(ctx context.Context, states []domain.State, limit, offset int) ([]*domain.Job, int, error) {
	var results []*domain.Job

	totalCount, err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollectionTitle)).
		Find(
			ctx,
			bson.M{
				"state": bson.M{"$in": states},
			},
			&results,
			mongodriver.Limit(limit), mongodriver.Offset(offset),
			mongodriver.Sort(bson.M{"last_updated": -1}),
		)
	return results, totalCount, err
}

// ClaimJob claims a pending job for processing.
func (m *Mongo) ClaimJob(ctx context.Context, pendingState, activeState domain.State) (*domain.Job, error) {
	var job domain.Job

	filter := bson.M{"state": pendingState}
	update := bson.M{
		"$set": bson.M{
			"state":        activeState,
			"last_updated": time.Now(),
		},
	}

	err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollectionTitle)).
		FindOneAndUpdate(ctx, filter, update, &job, mongodriver.ReturnDocument(options.After), mongodriver.Sort(bson.M{"last_updated": 1}))
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocumentFound) || errors.Is(err, mongo.ErrNoDocuments) {
			// If no pending jobs, no error.
			return nil, nil
		}
		return nil, err
	}

	return &job, nil
}

// GetJobsByConfig retrieves jobs based on the provided job configuration.
func (m *Mongo) GetJobsByConfig(ctx context.Context, jc *domain.JobConfig, limit, offset int) ([]*domain.Job, error) {
	return m.GetJobsByConfigAndState(ctx, jc, []domain.State{}, limit, offset)
}

// UpdateJob updates an existing migration job.
func (m *Mongo) UpdateJob(ctx context.Context, job *domain.Job) error {
	filter := bson.M{"_id": job.ID}
	update := bson.M{"$set": job}

	result, err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollectionTitle)).
		UpdateOne(ctx, filter, update)
	if err != nil {
		return appErrors.ErrInternalServerError
	}

	if result.MatchedCount == 0 {
		return appErrors.ErrJobNotFound
	}

	return nil
}

// UpdateJobState updates the state of a job and returns the updated job.
func (m *Mongo) UpdateJobState(ctx context.Context, jobID string, newState domain.State, lastUpdated time.Time) error {
	collectionName := m.ActualCollectionName(config.JobsCollectionTitle)

	filter := bson.M{"_id": jobID}
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
		return err
	}

	// Check if a document was found and updated
	if result.MatchedCount == 0 {
		return appErrors.ErrJobNotFound
	}

	return nil
}
