package mongo_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/ONSdigital/dis-migration-service/config"
	mongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	testMongoContainer "github.com/testcontainers/testcontainers-go/modules/mongodb"
)

const (
	mongoVersion = "4.4.8"
	database     = "test-migration-db"
)

var (
	sharedMongoServer *testMongoContainer.MongoDBContainer
	sharedConn        *mongoDriver.MongoConnection
	setupOnce         sync.Once
	setupErr          error
)

// setupSharedMongo initializes a shared MongoDB container for all tests
func setupSharedMongo(t *testing.T) (*mongoDriver.MongoConnection, error) {
	setupOnce.Do(func() {
		ctx := context.Background()
		var err error

		sharedMongoServer, err = testMongoContainer.Run(ctx, fmt.Sprintf("mongo:%s", mongoVersion))
		if err != nil {
			setupErr = fmt.Errorf("failed to start mongo server: %w", err)
			return
		}

		sharedConn, err = mongoDriver.Open(getMongoDriverConfig(sharedMongoServer, database, map[string]string{
			config.JobsCollectionTitle:   config.JobsCollectionName,
			config.EventsCollectionTitle: config.EventsCollectionName,
			config.TasksCollectionTitle:  config.TasksCollectionName,
		}))
		if err != nil {
			setupErr = fmt.Errorf("failed to open mongo connection: %w", err)
			return
		}
	})

	return sharedConn, setupErr
}
