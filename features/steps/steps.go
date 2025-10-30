package steps

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/ONSdigital/log.go/v2/log"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

func (c *MigrationComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I should receive a hello-world response$`, c.iShouldReceiveAHelloworldResponse)
	ctx.Step(`^mongodb is healthy$`, c.mongodbIsHealthy)
	ctx.Step(`^all its expected collections exist$`, c.allItsExpectedCollectionsExist)
	ctx.Step(`^the migration service is running$`, c.theMigrationServiceIsRunning)
	ctx.Step(`^mongodb stops running$`, c.mongodbStopsRunning)
}

func (c *MigrationComponent) iShouldReceiveAHelloworldResponse() error {
	responseBody := c.apiFeature.HTTPResponse.Body
	body, _ := io.ReadAll(responseBody)

	assert.Equal(c, `{"message":"Hello, World!"}`, strings.TrimSpace(string(body)))

	return c.StepError()
}

func (c *MigrationComponent) mongodbIsHealthy() error {
	err := c.mongoFeature.Client.Ping(context.Background(), nil)
	return err
}

func (c *MigrationComponent) theMigrationServiceIsRunning() error {
	if c.ServiceRunning {
		return nil // already started
	} else {
		return errors.New("expected the migration service to be running but it was not")
	}
}

func (c *MigrationComponent) allItsExpectedCollectionsExist() error {
	ctx := context.Background()

	err := c.mongoFeature.Client.Database(databaseName).CreateCollection(ctx, "jobs")
	if err != nil {
		return err
	}
	err = c.mongoFeature.Client.Database(databaseName).CreateCollection(ctx, "events")
	if err != nil {
		return err
	}
	err = c.mongoFeature.Client.Database(databaseName).CreateCollection(ctx, "tasks")
	return err
}

func (c *MigrationComponent) mongodbStopsRunning() error {
	err := c.MongoClient.Close(context.Background())
	if err != nil {
		log.Error(context.Background(), "error occurred while stopping the Mongo client", err)
	}
	return err
}
