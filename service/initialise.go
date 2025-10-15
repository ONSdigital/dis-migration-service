package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-migration-service/clients"
	clientMocks "github.com/ONSdigital/dis-migration-service/clients/mock"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/migrator"
	"github.com/ONSdigital/dis-migration-service/store"
	"github.com/ONSdigital/dis-migration-service/store/mock"
	redirectAPI "github.com/ONSdigital/dis-redirect-api/sdk/go"
	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-api-clients-go/v2/upload"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetAPI "github.com/ONSdigital/dp-dataset-api/sdk"

	"github.com/ONSdigital/log.go/v2/log"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	HealthCheck bool
	Init        Initialiser
	MongoDB     bool
	Migrator    bool
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		HealthCheck: false,
		Init:        initialiser,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

// GetHTTPServer creates an http server
func (e *ExternalServiceList) GetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := e.Init.DoGetHTTPServer(bindAddr, router)
	return s
}

// GetHealthCheck creates a healthcheck with versionInfo and sets teh HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	e.HealthCheck = true
	return hc, nil
}

// GetMongoDB returns a mongodb health client and dataset mongo object
func (e *ExternalServiceList) GetMongoDB(ctx context.Context, cfg config.MongoConfig) (store.MongoDB, error) {
	mongodb, err := e.Init.DoGetMongoDB(ctx, cfg)
	if err != nil {
		log.Error(ctx, "failed to initialise mongo", err)
		return nil, err
	}
	e.MongoDB = true
	return mongodb, nil
}

// GetMongoDB returns a mongodb health client and dataset mongo object
func (e *ExternalServiceList) GetMigrator(ctx context.Context, datastore store.Datastore, clientList *clients.ClientList) (migrator.Migrator, error) {
	mig, err := e.Init.DoGetMigrator(ctx, datastore, clientList)
	if err != nil {
		return nil, err
	}

	e.Migrator = true
	return mig, nil
}

// GetMongoDB returns a mongodb health client and dataset mongo object
func (e *ExternalServiceList) GetAppClients(ctx context.Context, cfg *config.Config) *clients.ClientList {
	return e.Init.DoGetAppClients(ctx, cfg)
}

// DoGetHTTPServer creates an HTTP Server with the provided bind address and router
func (e *Init) DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// DoGetHealthCheck creates a healthcheck with versionInfo
func (e *Init) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

// DoGetMongoDB returns a MongoDB
func (e *Init) DoGetMongoDB(ctx context.Context, cfg config.MongoConfig) (store.MongoDB, error) {
	// TODO: put in non-mocked MongoDB here
	// nolint:gocritic // helpful for non-mock implementation
	// mongodb := &mongo.Mongo{
	// 	MongoConfig: cfg,
	// }
	// if err := mongodb.Init(ctx); err != nil {
	// 	return nil, err
	// }
	log.Info(ctx, "listening to mongo db session")
	return &mock.MongoDBMock{
		GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
			return &domain.Job{
				ID:          jobID,
				LastUpdated: "test-time",
				State:       "submitted",
				Config: &domain.JobConfig{
					SourceID: "test-source-id",
					TargetID: "test-target-id",
					Type:     "test-type",
				},
			}, nil
		},
		CreateJobFunc: func(ctx context.Context, job *domain.Job) (*domain.Job, error) {
			return &domain.Job{
				ID:          "test-id",
				LastUpdated: "test-time",
				Config:      job.Config,
				State:       job.State,
			}, nil
		},
		CloseFunc: func(ctx context.Context) error { return nil },
	}, nil
}

// DoGetMigrator returns a Migrator
func (e *Init) DoGetMigrator(ctx context.Context, datastore store.Datastore, clientList *clients.ClientList) (migrator.Migrator, error) {
	mig := migrator.NewDefaultMigrator(datastore, clientList)
	log.Info(ctx, "migrator initialised")
	return mig, nil
}

// DoGetAppClients returns a set of app clients for the migration service
func (e *Init) DoGetAppClients(ctx context.Context, cfg *config.Config) *clients.ClientList {
	if cfg.EnableMockClients {
		return &clients.ClientList{
			DatasetAPI:    &clientMocks.DatasetAPIClientMock{},
			FilesAPI:      &clientMocks.FilesAPIClientMock{},
			RedirectAPI:   &clientMocks.RedirectAPIClientMock{},
			UploadService: &clientMocks.UploadServiceClientMock{},
			Zebedee:       &clientMocks.ZebedeeClientMock{},
		}
	}

	return &clients.ClientList{
		DatasetAPI:    datasetAPI.New(cfg.DatasetAPIURL),
		FilesAPI:      files.NewAPIClient(cfg.FilesAPIURL, cfg.ServiceAuthToken),
		RedirectAPI:   redirectAPI.NewClient(cfg.RedirectAPIURL),
		UploadService: upload.NewAPIClient(cfg.UploadServiceURL, cfg.ServiceAuthToken),
		Zebedee:       zebedee.New(cfg.ZebedeeURL),
	}
}
