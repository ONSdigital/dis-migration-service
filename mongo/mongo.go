package mongo

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	mongoHealth "github.com/ONSdigital/dp-mongodb/v3/health"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
)

// Mongo represents a mongo connection and health client
type Mongo struct {
	config.MongoConfig

	Connection         MongoConnection
	concreteConnection *mongodriver.MongoConnection // Concrete connection for health checks
	healthClient       *mongoHealth.CheckMongoClient
}

// Init returns an initialised Mongo object encapsulating a connection
// to the mongo server/cluster with the given configuration,
// and a health client to check the health of the mongo server/cluster
func (m *Mongo) Init(ctx context.Context) (err error) {
	m.concreteConnection, err = mongodriver.Open(&m.MongoDriverConfig)
	if err != nil {
		return err
	}

	// Wrap the concrete connection with our interface adapter
	m.Connection = NewMongoConnectionAdapter(m.concreteConnection)

	databaseCollectionBuilder := map[mongoHealth.Database][]mongoHealth.Collection{
		mongoHealth.Database(m.Database): {
			mongoHealth.Collection(m.ActualCollectionName(config.CountersCollectionTitle)),
			mongoHealth.Collection(m.ActualCollectionName(config.JobsCollectionTitle)),
			mongoHealth.Collection(m.ActualCollectionName(config.EventsCollectionTitle)),
			mongoHealth.Collection(m.ActualCollectionName(config.TasksCollectionTitle)),
		},
	}
	m.healthClient = mongoHealth.NewClientWithCollections(m.concreteConnection, databaseCollectionBuilder)

	return nil
}

// Close represents mongo session closing within the context deadline
func (m *Mongo) Close(ctx context.Context) error {
	if m.concreteConnection != nil {
		return m.concreteConnection.Close(ctx)
	}
	return nil
}

// Checker is called by the healthcheck library to check the health
// state of this mongoDB instance
func (m *Mongo) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	return m.healthClient.Checker(ctx, state)
}

// getConnection returns the connection to use for operations.
// In production, this is the adapter wrapping the real connection.
// In tests, this is the injected mock connection.
func (m *Mongo) getConnection() MongoConnection {
	return m.Connection
}

// SetConnection allows tests to inject a mock connection
func (m *Mongo) SetConnection(conn MongoConnection) {
	m.Connection = conn
}
