package mongo

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/config"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"

	mongohealth "github.com/ONSdigital/dp-mongodb/v3/health"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
)

type Mongo struct {
	config.MongoConfig

	Connection   *mongodriver.MongoConnection
	healthClient *mongohealth.CheckMongoClient
}

// Init returns an initialised Mongo object encapsulating a connection to the mongo server/cluster with the given configuration,
// and a health client to check the health of the mongo server/cluster
func (m *Mongo) Init(ctx context.Context) (err error) {
	m.Connection, err = mongodriver.Open(&m.MongoDriverConfig)
	if err != nil {
		return err
	}

	databaseCollectionBuilder := map[mongohealth.Database][]mongohealth.Collection{
		mongohealth.Database(m.Database): {
			mongohealth.Collection(m.ActualCollectionName(config.JobsCollectionTitle)),
			mongohealth.Collection(m.ActualCollectionName(config.EventsCollectionTitle)),
			mongohealth.Collection(m.ActualCollectionName(config.TasksCollectionTitle)),
		},
	}
	m.healthClient = mongohealth.NewClientWithCollections(m.Connection, databaseCollectionBuilder)

	return nil
}

// Close represents mongo session closing within the context deadline
func (m *Mongo) Close(ctx context.Context) error {
	return m.Connection.Close(ctx)
}

// Checker is called by the healthcheck library to check the health state of this mongoDB instance
func (m *Mongo) Checker(ctx context.Context, state *health.CheckState) error {
	return m.healthClient.Checker(ctx, state)
}
