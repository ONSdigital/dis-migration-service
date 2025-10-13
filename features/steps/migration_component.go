package steps

import (
	"context"
	"fmt"
	"github.com/ONSdigital/dp-component-test/utils"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"net/http"
	"strconv"
	"time"

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
)

type MigrationComponent struct {
	componenttest.ErrorFeature
	svcList        *service.ExternalServiceList
	svc            *service.Service
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
				ClusterEndpoint: c.mongoFeature.Server.URI(),
				Database:        utils.RandomDatabase(),
				Collections:     c.Config.Collections,
				ConnectTimeout:  c.Config.ConnectTimeout,
				QueryTimeout:    c.Config.QueryTimeout,
			},
		}}

	if err := mongodb.Init(context.Background()); err != nil {
		return &MigrationComponent{}, err
	}

	c.MongoClient = mongodb

	initMock := &mock.InitialiserMock{
		DoGetHealthCheckFunc: c.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:  c.DoGetHTTPServer,
		DoGetMongoDBFunc:     c.DoGetMongoDB,
	}

	c.apiFeature = componenttest.NewAPIFeature(c.InitialiseService)
	c.Config.HealthCheckInterval = 1 * time.Second
	c.Config.HealthCheckCriticalTimeout = 3 * time.Second
	c.Config.BindAddr = "localhost:0"
	c.StartTime = time.Now()
	c.svcList = service.NewServiceList(initMock)

	c.HTTPServer = &http.Server{ReadHeaderTimeout: 3 * time.Second}
	c.svc, err = service.Run(context.Background(), c.Config, c.svcList, "1", "", "", c.errorChan)
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

func (c *MigrationComponent) Reset() *MigrationComponent {
	c.MongoClient.Database = utils.RandomDatabase()
	c.apiFeature.Reset()
	return c
}

func (c *MigrationComponent) Close() error {
	if c.svc != nil && c.ServiceRunning {
		c.svc.Close(context.Background())
		c.ServiceRunning = false
	}
	return nil
}

func (c *MigrationComponent) InitialiseService() (http.Handler, error) {
	var err error
	c.svc, err = service.Run(context.Background(), c.Config, c.svcList, "1", "", "", c.errorChan)
	if err != nil {
		return nil, err
	}

	c.ServiceRunning = true
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
	c.HTTPServer.Addr = bindAddr
	c.HTTPServer.Handler = router
	return c.HTTPServer
}

func (c *MigrationComponent) DoGetMongoDB(context.Context, config.MongoConfig) (store.MongoDB, error) {
	return c.MongoClient, nil
}
