package steps

import (
	"context"

	"github.com/ONSdigital/log.go/v2/log"

	"github.com/cucumber/godog"
)

func (c *MigrationComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^mongodb is healthy$`, c.mongodbIsHealthy)
	ctx.Step(`^all its expected collections exist$`, c.allItsExpectedCollectionsExist)
	ctx.Step(`^the migration service is running$`, c.restartMigrationService)
	ctx.Step(`^mongodb stops running$`, c.mongodbStopsRunning)
	ctx.Step(`^a get page data request to zebedee for "([^"]*)" returns a page of type "([^"]*)" with status (\d+)$`, c.getPageDataRequestToZebedeeForReturnsAPageOfTypeWithStatus)
	ctx.Step(`^a get page data request to zebedee for "([^"]*)" returns with status (\d+) and payload:$`, c.getPageDataRequestToZebedeeForReturnsWithPayload)
	ctx.Step(`^a get dataset request to the dataset API for "([^"]*)" returns with status (\d+)$`, c.getDatasetRequestToDatasetAPIForReturnsWithStatus)
}

func (c *MigrationComponent) mongodbIsHealthy() error {
	err := c.mongoFeature.Client.Ping(context.Background(), nil)
	return err
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

func (c *MigrationComponent) getPageDataRequestToZebedeeForReturnsAPageOfTypeWithStatus(url, pageType string, statusCode int) error {
	c.FakeAPIRouter.setJSONResponseForGetPageData(url, pageType, statusCode)
	return nil
}

func (c *MigrationComponent) getPageDataRequestToZebedeeForReturnsWithPayload(url string, statusCode int, payload *godog.DocString) error {
	c.FakeAPIRouter.setFullJSONResponseForGetPageData(url, statusCode, payload.Content)
	return nil
}

func (c *MigrationComponent) getDatasetRequestToDatasetAPIForReturnsWithStatus(id string, statusCode int) error {
	c.FakeAPIRouter.setJSONResponseForGetDataset(id, statusCode)
	return nil
}

func (c *MigrationComponent) restartMigrationService() error {
	return c.Restart()
}
