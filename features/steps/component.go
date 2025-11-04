package steps

import (
	"context"
	"net/http"
	"time"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	clientMocks "github.com/ONSdigital/dis-migration-service/clients/mock"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/migrator"
	migratorMock "github.com/ONSdigital/dis-migration-service/migrator/mock"
	"github.com/ONSdigital/dis-migration-service/service"
	"github.com/ONSdigital/dis-migration-service/service/mock"
	"github.com/ONSdigital/dis-migration-service/store"
	storeMock "github.com/ONSdigital/dis-migration-service/store/mock"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

type Component struct {
	componenttest.ErrorFeature
	svcList        *service.ExternalServiceList
	svc            *service.Service
	errorChan      chan error
	Config         *config.Config
	HTTPServer     *http.Server
	ServiceRunning bool
	apiFeature     *componenttest.APIFeature
}

func NewComponent() (*Component, error) {
	c := &Component{
		HTTPServer:     &http.Server{ReadHeaderTimeout: 3 * time.Second},
		errorChan:      make(chan error),
		ServiceRunning: false,
	}

	var err error

	c.Config, err = config.Get()
	if err != nil {
		return nil, err
	}

	initMock := &mock.InitialiserMock{
		DoGetHealthCheckFunc: c.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:  c.DoGetHTTPServer,
		DoGetMigratorFunc:    c.DoGetMigrator,
		DoGetMongoDBFunc:     c.DoGetMongoDB,
		DoGetAppClientsFunc:  c.DoGetAppClients,
	}

	c.svcList = service.NewServiceList(initMock)

	c.apiFeature = componenttest.NewAPIFeature(c.InitialiseService)

	return c, nil
}

func (c *Component) Reset() *Component {
	c.apiFeature.Reset()
	return c
}

func (c *Component) Close() error {
	if c.svc != nil && c.ServiceRunning {
		c.svc.Close(context.Background())
		c.ServiceRunning = false
	}
	return nil
}

func (c *Component) InitialiseService() (http.Handler, error) {
	var err error
	svc := service.New(c.Config, c.svcList)
	err = svc.Run(context.Background(), "1", "", "", c.errorChan)
	if err != nil {
		return nil, err
	}

	c.ServiceRunning = true
	return c.HTTPServer.Handler, nil
}

func (c *Component) DoGetHealthcheckOk(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
	return &mock.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		StartFunc:    func(ctx context.Context) {},
		StopFunc:     func() {},
	}, nil
}

func (c *Component) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	c.HTTPServer.Addr = bindAddr
	c.HTTPServer.Handler = router
	return c.HTTPServer
}

func (c *Component) DoGetMongoDB(ctx context.Context, cfg config.MongoConfig) (store.MongoDB, error) {
	return &storeMock.MongoDBMock{
		GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
			return &domain.Job{
				ID:          jobID,
				LastUpdated: time.Now(),
				State:       "submitted",
				Config: &domain.JobConfig{
					SourceID: "test-source-id",
					TargetID: "test-target-id",
					Type:     "test-type",
				},
			}, nil
		},
		CreateJobFunc: func(ctx context.Context, job *domain.Job) error {
			return nil
		},
		CloseFunc: func(ctx context.Context) error { return nil },
	}, nil
}

func (c *Component) DoGetMigrator(ctx context.Context, jobService application.JobService, clientList *clients.ClientList) (migrator.Migrator, error) {
	mig := &migratorMock.MigratorMock{
		MigrateFunc: func(ctx context.Context, job *domain.Job) {},
	}

	return mig, nil
}

func (c *Component) DoGetAppClients(ctx context.Context, cfg *config.Config) *clients.ClientList {
	return &clients.ClientList{
		DatasetAPI:    &clientMocks.DatasetAPIClientMock{},
		FilesAPI:      &clientMocks.FilesAPIClientMock{},
		RedirectAPI:   &clientMocks.RedirectAPIClientMock{},
		UploadService: &clientMocks.UploadServiceClientMock{},
		Zebedee:       &clientMocks.ZebedeeClientMock{},
	}
}
