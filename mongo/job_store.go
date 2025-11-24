package mongo

import (
	"context"
	"errors"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
		if errors.Is(err, mongodriver.ErrNoDocumentFound) || errors.Is(err, mongo.ErrNoDocuments) {
			return nil, appErrors.ErrJobNotFound
		}
		return nil, err
	}
	return &job, nil
}

// GetJobsByConfigAndState retrieves jobs based on the provided job
// configuration and states.
func (m *Mongo) GetJobsByConfigAndState(ctx context.Context, jc *domain.JobConfig, stateFilter []domain.JobState, offset, limit int) ([]*domain.Job, error) {
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
			mongodriver.Offset(offset), mongodriver.Limit(limit),
		)

	return results, err
}

// GetJobsByConfig retrieves jobs based on the provided job configuration.
func (m *Mongo) GetJobsByConfig(ctx context.Context, jc *domain.JobConfig, offset, limit int) ([]*domain.Job, error) {
	return m.GetJobsByConfigAndState(ctx, jc, []domain.JobState{}, offset, limit)
}
