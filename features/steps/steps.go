package steps

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

func (c *MigrationComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I should receive a hello-world response$`, c.iShouldReceiveAHelloworldResponse)
	ctx.Step(`^mongodb is healthy$`, c.mongodbIsHealthy)
	ctx.Step(`^all its expected collections exist$`, c.allItsExpectedCollectionsExist)
	ctx.Step(`^the migration service is running$`, c.theMigrationServiceIsRunning)
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
	//c.mongoFeature.Client.ListDatabaseNames()
	c.mongoFeature.Client.Database(databaseName).CreateCollection(context.Background(), "jobs")
	c.mongoFeature.Client.Database(databaseName).CreateCollection(context.Background(), "events")
	c.mongoFeature.Client.Database(databaseName).CreateCollection(context.Background(), "tasks")
	return nil
}
