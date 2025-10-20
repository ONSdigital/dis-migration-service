package steps

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-component-test/utils"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"

	"github.com/ONSdigital/dis-migration-service/mongo"
	"github.com/ONSdigital/dis-migration-service/store"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/service"
	"github.com/ONSdigital/dis-migration-service/service/mock"
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
	svcList        *service.ExternalServiceList
	serviceSvc     *service.Service
	errorChan      chan error
	Config         *config.Config
	HTTPServer     *http.Server
	ServiceRunning bool
	apiFeature     *componenttest.APIFeature
	MongoClient    *mongo.Mongo
	mongoFeature   *componenttest.MongoFeature
	StartTime      time.Time
}

func NewMigrationComponent(mongoFeat *componenttest.MongoFeature) (*MigrationComponent, error) {
	c := &MigrationComponent{
		HTTPServer:     &http.Server{ReadHeaderTimeout: 3 * time.Second},
		errorChan:      make(chan error),
		ServiceRunning: false,
		mongoFeature:   mongoFeat,
	}

	var err error

	c.Config, err = config.Get()
	if err != nil {
		return &MigrationComponent{}, fmt.Errorf("failed to get config: %w", err)
	}

	mongodb := &mongo.Mongo{
		MongoConfig: config.MongoConfig{
			MongoDriverConfig: mongodriver.MongoDriverConfig{
				ClusterEndpoint: mongoFeat.Server.URI(),
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
	c.serviceSvc = service.New(c.Config, c.svcList)
	err = c.serviceSvc.Run(context.Background(), "1", "", "", c.errorChan)
	if err != nil {
		return &MigrationComponent{}, err
	}
	c.ServiceRunning = true

	return c, nil
}

func (c *MigrationComponent) InitAPIFeature() *componenttest.APIFeature {
	c.apiFeature = componenttest.NewAPIFeature(c.InitialiseService)

	return c.apiFeature
}

func (c *MigrationComponent) Reset() error {
	c.MongoClient.Database = utils.RandomDatabase()
	c.apiFeature.Reset()

	return nil
}

func (c *MigrationComponent) Close() error {
	if c.serviceSvc != nil && c.ServiceRunning {
		if err := c.serviceSvc.Close(context.Background()); err != nil {
			return err
		}
		c.ServiceRunning = false
	}

	return nil
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

func (c *MigrationComponent) DoGetMongoDB(context.Context, config.MongoConfig) (store.MongoDB, error) {
	return c.MongoClient, nil
}
