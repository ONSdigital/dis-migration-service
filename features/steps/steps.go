package steps

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ONSdigital/log.go/v2/log"

	"github.com/cucumber/godog"
)

func (c *MigrationComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^mongodb is healthy$`, c.mongodbIsHealthy)
	ctx.Step(`^all its expected collections exist$`, c.allItsExpectedCollectionsExist)
	ctx.Step(`^the migration service is running$`, c.restartMigrationService)
	ctx.Step(`^mongodb stops running$`, c.mongodbStopsRunning)
	ctx.Step(`^a get page data request to zebedee for "([^"]*)" returns with status (\d+) and payload:$`, c.getPageDataRequestToZebedeeForReturnsWithPayload)
	ctx.Step(`^a get dataset request to the dataset API for "([^"]*)" returns with status (\d+)$`, c.getDatasetRequestToDatasetAPIForReturnsWithStatus)
	ctx.Step(`^the Dataset API responds successfully to create dataset requests$`, c.datasetAPIrespondsSuccessfullyToCreateDatasetRequests)

	// Slack notification assertions
	ctx.Step(`^(\d+) Slack info notifications? should have been sent$`, c.slackInfoNotificationsShouldHaveBeenSent)
	ctx.Step(`^(\d+) Slack warning notifications? should have been sent$`, c.slackWarningNotificationsShouldHaveBeenSent)
	ctx.Step(`^(\d+) Slack alarm notifications? should have been sent$`, c.slackAlarmNotificationsShouldHaveBeenSent)
	ctx.Step(`^a Slack info notification with summary "([^"]*)" should have been sent$`, c.slackInfoNotificationWithSummaryShouldHaveBeenSent)
	ctx.Step(`^a Slack warning notification with summary "([^"]*)" should have been sent$`, c.slackWarningNotificationWithSummaryShouldHaveBeenSent)
	ctx.Step(`^a Slack alarm notification with summary "([^"]*)" should have been sent$`, c.slackAlarmNotificationWithSummaryShouldHaveBeenSent)
	ctx.Step(`^a Slack info notification with detail "([^"]*)" = "([^"]*)" should have been sent$`, c.slackInfoNotificationWithDetailShouldHaveBeenSent)
	ctx.Step(`^a Slack info notification with detail "([^"]*)" = (\d+) should have been sent$`, c.slackInfoNotificationWithDetailIntShouldHaveBeenSent)
	ctx.Step(`^no Slack notifications should have been sent$`, c.noSlackNotificationsShouldHaveBeenSent)
	ctx.Step(`^the Slack notification history is cleared$`, c.slackNotificationHistoryIsCleared)
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

func (c *MigrationComponent) datasetAPIrespondsSuccessfullyToCreateDatasetRequests() {
	c.FakeAPIRouter.setJSONResponseForCreateDataset(http.StatusCreated)
}

func (c *MigrationComponent) slackInfoNotificationsShouldHaveBeenSent(count int) error {
	return c.AssertSlackInfoCalled(count)
}

func (c *MigrationComponent) slackWarningNotificationsShouldHaveBeenSent(count int) error {
	return c.AssertSlackWarningCalled(count)
}

func (c *MigrationComponent) slackAlarmNotificationsShouldHaveBeenSent(count int) error {
	return c.AssertSlackAlarmCalled(count)
}

func (c *MigrationComponent) slackInfoNotificationWithSummaryShouldHaveBeenSent(summary string) error {
	return c.AssertSlackInfoCalledWithSummary(summary)
}

func (c *MigrationComponent) slackWarningNotificationWithSummaryShouldHaveBeenSent(summary string) error {
	return c.AssertSlackWarningCalledWithSummary(summary)
}

func (c *MigrationComponent) slackAlarmNotificationWithSummaryShouldHaveBeenSent(summary string) error {
	return c.AssertSlackAlarmCalledWithSummary(summary)
}

func (c *MigrationComponent) slackInfoNotificationWithDetailShouldHaveBeenSent(key, value string) error {
	return c.AssertSlackInfoCalledWithDetails(key, value)
}

func (c *MigrationComponent) slackInfoNotificationWithDetailIntShouldHaveBeenSent(key string, value int) error {
	return c.AssertSlackInfoCalledWithDetails(key, value)
}

func (c *MigrationComponent) noSlackNotificationsShouldHaveBeenSent() error {
	infoCount := len(c.MockSlackClient.SendInfoCalls())
	warningCount := len(c.MockSlackClient.SendWarningCalls())
	alarmCount := len(c.MockSlackClient.SendAlarmCalls())

	totalCount := infoCount + warningCount + alarmCount
	if totalCount != 0 {
		return fmt.Errorf("expected no Slack notifications to be sent, but %d were sent (info: %d, warning: %d, alarm: %d)",
			totalCount, infoCount, warningCount, alarmCount)
	}
	return nil
}

func (c *MigrationComponent) slackNotificationHistoryIsCleared() error {
	c.ResetMockSlackClient()
	return nil
}
