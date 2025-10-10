package steps

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/service"
	"github.com/ONSdigital/dis-migration-service/service/mock"
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

func (c *MigrationComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	c.apiFeature.RegisterSteps(ctx)

	ctx.Step(`^I should receive a hello-world response$`, c.iShouldReceiveAHelloworldResponse)
	ctx.Step(`^mongodb is healthy$`, mongodbIsHealthy)
	ctx.Step(`^the migration service is running$`, c.theMigrationServiceIsRunning)
}

func (c *MigrationComponent) iShouldReceiveAHelloworldResponse() error {
	responseBody := c.apiFeature.HTTPResponse.Body
	body, _ := io.ReadAll(responseBody)

	assert.Equal(c, `{"message":"Hello, World!"}`, strings.TrimSpace(string(body)))

	return c.StepError()
}

func mongodbIsHealthy() error {
	return godog.ErrPending
}

func (c *MigrationComponent) theMigrationServiceIsRunning() error {
	if c.ServiceRunning {
		return nil // already started
	}

	var err error

	c.Config, err = config.Get()
	if err != nil {
		return err
	}

	c.Config.MongoConfig.ClusterEndpoint = c.mongoFeature.Server.URI()

	initMock := &mock.InitialiserMock{
		DoGetHealthCheckFunc: c.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:  c.DoGetHTTPServer,
		DoGetMongoDBFunc:     c.DoGetMongoDB,
	}

	c.Config.HealthCheckInterval = 1 * time.Second
	c.Config.HealthCheckCriticalTimeout = 3 * time.Second
	c.Config.BindAddr = "localhost:0"
	c.StartTime = time.Now()
	c.svcList = service.NewServiceList(initMock)

	c.HTTPServer = &http.Server{ReadHeaderTimeout: 3 * time.Second}
	c.svc, err = service.Run(context.Background(), c.Config, c.svcList, "1", "", "", c.errorChan)
	if err != nil {
		return err
	}

	c.ServiceRunning = true
	return nil
}
