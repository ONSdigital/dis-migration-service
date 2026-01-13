package steps

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/migrator"
	"github.com/ONSdigital/dis-migration-service/mongo"
	"github.com/ONSdigital/dis-migration-service/service"
	"github.com/ONSdigital/dis-migration-service/service/mock"
	"github.com/ONSdigital/dis-migration-service/slack"
	slackMocks "github.com/ONSdigital/dis-migration-service/slack/mocks"
	"github.com/ONSdigital/dis-migration-service/store"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	gitCommitHash = "6584b786caac36b6214ffe04bf62f058d4021538"
	appVersion    = "v1.2.3"
	databaseName  = "testing"
)

type MigrationComponent struct {
	componenttest.ErrorFeature
	svcList                 *service.ExternalServiceList
	svc                     *service.Service
	errorChan               chan error
	Config                  *config.Config
	HTTPServer              *http.Server
	ServiceRunning          bool
	apiFeature              *componenttest.APIFeature
	MongoClient             *mongo.Mongo
	mongoFeature            *componenttest.MongoFeature
	authFeature             *componenttest.AuthorizationFeature
	StartTime               time.Time
	FakeAPIRouter           *FakeAPI
	AuthorisationMiddleware authorisation.Middleware
	MockSlackClient         *slackMocks.ClienterMock
	Migrator                migrator.Migrator
}

func NewMigrationComponent(mongoFeat *componenttest.MongoFeature, authFeat *componenttest.AuthorizationFeature) (*MigrationComponent, error) {
	c := &MigrationComponent{
		HTTPServer:     &http.Server{ReadHeaderTimeout: 3 * time.Second},
		errorChan:      make(chan error),
		ServiceRunning: false,
		mongoFeature:   mongoFeat,
		authFeature:    authFeat,
	}

	var err error

	c.Config, err = config.Get()
	if err != nil {
		return &MigrationComponent{}, fmt.Errorf("failed to get config: %w", err)
	}

	c.Config.MigratorPollInterval = 500 * time.Millisecond

	mongoURI, err := mongoFeat.GetConnectionString()
	if err != nil {
		panic(err)
	}

	// Extract host:port from the MongoDB URI
	parsedURI, err := url.Parse(mongoURI)
	if err != nil {
		return nil, err
	}
	hostPort := parsedURI.Host

	mongodb := &mongo.Mongo{
		MongoConfig: config.MongoConfig{
			MongoDriverConfig: mongodriver.MongoDriverConfig{
				ClusterEndpoint: hostPort,
				Database:        databaseName,
				Collections:     c.Config.Collections,
				ConnectTimeout:  c.Config.ConnectTimeout,
				QueryTimeout:    c.Config.QueryTimeout,
			},
		}}

	ctx := context.Background()
	if dbErr := mongodb.Init(ctx); dbErr != nil {
		return nil, fmt.Errorf("failed to initialise mongo DB: %w", dbErr)
	}

	c.MongoClient = mongodb

	c.FakeAPIRouter = NewFakeAPI()
	c.Config.ZebedeeURL = c.FakeAPIRouter.fakeHTTP.Server.URL
	c.Config.DatasetAPIURL = c.FakeAPIRouter.fakeHTTP.Server.URL

	// Initialize mock Slack client
	c.MockSlackClient = &slackMocks.ClienterMock{
		SendInfoFunc: func(ctx context.Context, summary string, details slack.SlackDetails) error {
			log.Info(ctx, "mock slack: info notification", log.Data{
				"summary": summary,
				"details": details,
			})
			return nil
		},
		SendWarningFunc: func(ctx context.Context, summary string, details slack.SlackDetails) error {
			log.Info(ctx, "mock slack: warning notification", log.Data{
				"summary": summary,
				"details": details,
			})
			return nil
		},
		SendAlarmFunc: func(ctx context.Context, summary string, err error, details slack.SlackDetails) error {
			log.Info(ctx, "mock slack: alarm notification", log.Data{
				"summary": summary,
				"error":   err,
				"details": details,
			})
			return nil
		},
	}

	initMock := &mock.InitialiserMock{
		DoGetHealthCheckFunc:             c.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:              c.DoGetHTTPServer,
		DoGetMongoDBFunc:                 c.DoGetMongoDB,
		DoGetSlackClientFunc:             c.DoGetSlackClient,
		DoGetMigratorFunc:                c.DoGetMigrator,
		DoGetAppClientsFunc:              c.DoGetAppClients,
		DoGetAuthorisationMiddlewareFunc: c.DoGetAuthorisationMiddleware,
	}

	c.Config.HealthCheckInterval = 1 * time.Second
	c.Config.HealthCheckCriticalTimeout = 3 * time.Second
	c.Config.BindAddr = "localhost:0"
	c.Config.AuthConfig.PermissionsAPIURL = c.authFeature.FakePermissionsAPI.URL()
	c.StartTime = time.Now()
	c.svcList = service.NewServiceList(initMock)
	c.svc = service.New(c.Config, c.svcList)

	return c, nil
}

func (c *MigrationComponent) InitAPIFeature() *componenttest.APIFeature {
	c.apiFeature = componenttest.NewAPIFeature(c.InitialiseService)

	return c.apiFeature
}

func (c *MigrationComponent) Reset() error {
	return nil
}

func (c *MigrationComponent) Close() error {
	if c.svc != nil && c.ServiceRunning {
		if err := c.svc.Close(context.Background()); err != nil {
			return err
		}
		c.ServiceRunning = false
	}

	return nil
}

func (c *MigrationComponent) Start() error {
	if c.svc != nil && !c.ServiceRunning {
		err := c.svc.Run(context.Background(), "1", "", "", c.errorChan)
		if err != nil {
			return err
		}
		c.ServiceRunning = true
	}

	return nil
}

func (c *MigrationComponent) SeedDatabase() error {
	for key, value := range c.Config.Collections {
		log.Info(context.Background(), "creating collection", log.Data{"wellknown": key, "actual": value})
		cmd := bson.D{{Key: "create", Value: value}}
		err := c.MongoClient.Connection.RunCommand(context.Background(), cmd)
		if err != nil {
			return fmt.Errorf("failed to create collection %s: %w", value, err)
		}
	}

	colls, err := c.MongoClient.Connection.ListCollectionsFor(context.Background(), c.MongoClient.Database)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}
	log.Info(context.Background(), "existing collections after seeding", log.Data{"collections": colls, "database": c.MongoClient.Database})
	return nil
}

func (c *MigrationComponent) Restart() error {
	err := c.Close()
	if err != nil {
		return err
	}
	return c.Start()
}

func (c *MigrationComponent) InitialiseService() (http.Handler, error) {
	return c.HTTPServer.Handler, nil
}

func (c *MigrationComponent) DoGetHealthcheckOk(cfg *config.Config, _, _, _ string) (service.HealthChecker, error) {
	componentBuildTime := strconv.Itoa(int(time.Now().Unix()))
	versionInfo, err := healthcheck.NewVersionInfo(componentBuildTime, gitCommitHash, appVersion)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

func (c *MigrationComponent) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	c.HTTPServer = &http.Server{
		ReadHeaderTimeout: 3 * time.Second,
		Addr:              bindAddr,
		Handler:           router,
	}
	return c.HTTPServer
}

func (c *MigrationComponent) DoGetMongoDB(ctx context.Context, cfg config.MongoConfig) (store.MongoDB, error) {
	return c.MongoClient, nil
}

func (c *MigrationComponent) DoGetSlackClient(ctx context.Context, cfg *config.Config) (slack.Clienter, error) {
	log.Info(ctx, "returning mock slack client for component testing")
	return c.MockSlackClient, nil
}

func (c *MigrationComponent) DoGetMigrator(ctx context.Context, cfg *config.Config, jobService application.JobService, clientList *clients.ClientList, slackClient slack.Clienter) (migrator.Migrator, error) {
	mig := migrator.NewDefaultMigrator(cfg, jobService, clientList, slackClient)
	c.Migrator = mig
	return mig, nil
}

func (c *MigrationComponent) DoGetAppClients(ctx context.Context, cfg *config.Config) *clients.ClientList {
	init := service.Init{}

	return init.DoGetAppClients(ctx, cfg)
}

func (c *MigrationComponent) DoGetAuthorisationMiddleware(ctx context.Context, cfg *authorisation.Config) (authorisation.Middleware, error) {
	middleware, err := authorisation.NewMiddlewareFromConfig(ctx, cfg, cfg.JWTVerificationPublicKeys)
	if err != nil {
		return nil, err
	}

	c.AuthorisationMiddleware = middleware
	return c.AuthorisationMiddleware, nil
}

func (c *MigrationComponent) GetMockSlackClient() *slackMocks.ClienterMock {
	return c.MockSlackClient
}

// AssertSlackInfoCalled asserts that SendInfo was called on the mock Slack client
func (c *MigrationComponent) AssertSlackInfoCalled(expectedCount int) error {
	actualCount := len(c.MockSlackClient.SendInfoCalls())
	if actualCount != expectedCount {
		return fmt.Errorf("expected SendInfo to be called %d times, but was called %d times", expectedCount, actualCount)
	}
	return nil
}

// AssertSlackInfoCalledWithSummary asserts that SendInfo was called with a specific summary
func (c *MigrationComponent) AssertSlackInfoCalledWithSummary(expectedSummary string) error {
	calls := c.MockSlackClient.SendInfoCalls()
	for _, call := range calls {
		if call.Summary == expectedSummary {
			return nil
		}
	}
	return fmt.Errorf("expected SendInfo to be called with summary %q, but it was not found", expectedSummary)
}

// AssertSlackInfoCalledWithDetails asserts that SendInfo was called with specific details
func (c *MigrationComponent) AssertSlackInfoCalledWithDetails(expectedKey string, expectedValue interface{}) error {
	calls := c.MockSlackClient.SendInfoCalls()
	for _, call := range calls {
		if val, ok := call.Details[expectedKey]; ok {
			if val == expectedValue {
				return nil
			}
		}
	}
	return fmt.Errorf("expected SendInfo to be called with details %s=%v, but it was not found", expectedKey, expectedValue)
}

// AssertSlackWarningCalled asserts that SendWarning was called on the mock Slack client
func (c *MigrationComponent) AssertSlackWarningCalled(expectedCount int) error {
	actualCount := len(c.MockSlackClient.SendWarningCalls())
	if actualCount != expectedCount {
		return fmt.Errorf("expected SendWarning to be called %d times, but was called %d times", expectedCount, actualCount)
	}
	return nil
}

// AssertSlackWarningCalledWithSummary asserts that SendWarning was called with a specific summary
func (c *MigrationComponent) AssertSlackWarningCalledWithSummary(expectedSummary string) error {
	calls := c.MockSlackClient.SendWarningCalls()
	for _, call := range calls {
		if call.Summary == expectedSummary {
			return nil
		}
	}
	return fmt.Errorf("expected SendWarning to be called with summary %q, but it was not found", expectedSummary)
}

// AssertSlackAlarmCalled asserts that SendAlarm was called on the mock Slack client
func (c *MigrationComponent) AssertSlackAlarmCalled(expectedCount int) error {
	actualCount := len(c.MockSlackClient.SendAlarmCalls())
	if actualCount != expectedCount {
		return fmt.Errorf("expected SendAlarm to be called %d times, but was called %d times", expectedCount, actualCount)
	}
	return nil
}

// AssertSlackAlarmCalledWithSummary asserts that SendAlarm was called with a specific summary
func (c *MigrationComponent) AssertSlackAlarmCalledWithSummary(expectedSummary string) error {
	calls := c.MockSlackClient.SendAlarmCalls()
	for _, call := range calls {
		if call.Summary == expectedSummary {
			return nil
		}
	}
	return fmt.Errorf("expected SendAlarm to be called with summary %q, but it was not found", expectedSummary)
}

// ResetMockSlackClient resets the mock Slack client call history
func (c *MigrationComponent) ResetMockSlackClient() {
	// Get all existing calls to clear them
	// This maintains the same mock instance that the migrator holds a reference to
	_ = c.MockSlackClient.SendInfoCalls()
	_ = c.MockSlackClient.SendWarningCalls()
	_ = c.MockSlackClient.SendAlarmCalls()

	// The moq-generated mock doesn't provide a Reset method,
	// so we need to recreate the mock while maintaining the reference
	// by keeping the same pointer but replacing its contents
	*c.MockSlackClient = slackMocks.ClienterMock{
		SendInfoFunc:    c.MockSlackClient.SendInfoFunc,
		SendWarningFunc: c.MockSlackClient.SendWarningFunc,
		SendAlarmFunc:   c.MockSlackClient.SendAlarmFunc,
	}
}
