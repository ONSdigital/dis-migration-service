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
	"github.com/ONSdigital/dis-migration-service/service/mock"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/ONSdigital/dis-migration-service/migrator"

	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"

	"github.com/ONSdigital/dis-migration-service/mongo"
	"github.com/ONSdigital/dis-migration-service/store"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/service"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
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

	c.Config.MigratorPollInterval = 2 * time.Second

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

	initMock := &mock.InitialiserMock{
		DoGetHealthCheckFunc:             c.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:              c.DoGetHTTPServer,
		DoGetMongoDBFunc:                 c.DoGetMongoDB,
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
		log.Info(context.Background(), "closing migration service")
		if err := c.svc.Close(context.Background()); err != nil {
			return err
		}
		c.ServiceRunning = false
	}

	return nil
}

func (c *MigrationComponent) Start() error {
	if c.svc != nil && !c.ServiceRunning {
		log.Info(context.Background(), "starting migration service")
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

func (c *MigrationComponent) DoGetMigrator(ctx context.Context, cfg *config.Config, jobService application.JobService, clientList *clients.ClientList) (migrator.Migrator, error) {
	mig := migrator.NewDefaultMigrator(cfg, jobService, clientList)
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
