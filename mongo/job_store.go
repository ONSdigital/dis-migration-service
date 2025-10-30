package mongo

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *Mongo) CreateJob(ctx context.Context, job *domain.Job) error {
	_, err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollectionTitle)).InsertOne(ctx, job)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mongo) GetJob(ctx context.Context, jobID string) (*domain.Job, error) {
	var job domain.Job
	// TODO: Implement this function
	err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollectionTitle)).FindOne(ctx, bson.M{"_id": jobID}, job)

	return &job, err
}

func (m *Mongo) GetJobsByConfigAndState(ctx context.Context, jc *domain.JobConfig, stateFilter []domain.JobState, offset, limit int) ([]*domain.Job, error) {
	var results []*domain.Job

	_, err := m.Connection.Collection(m.ActualCollectionName(config.JobsCollectionTitle)).
		Find(
			ctx,
			bson.M{
				"config": &jc,
				"state":  bson.M{"$in": stateFilter},
			},
			&results,
			mongodriver.Offset(offset), mongodriver.Limit(limit),
		)

	return results, err
}

func (m *Mongo) GetJobsByConfig(ctx context.Context, jc *domain.JobConfig, offset, limit int) ([]*domain.Job, error) {
	return m.GetJobsByConfigAndState(ctx, jc, []domain.JobState{}, offset, limit)
}
