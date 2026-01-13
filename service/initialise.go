package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/cache"
	"github.com/ONSdigital/dis-migration-service/clients"
	clientMocks "github.com/ONSdigital/dis-migration-service/clients/mock"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/migrator"
	"github.com/ONSdigital/dis-migration-service/mongo"
	"github.com/ONSdigital/dis-migration-service/slack"
	"github.com/ONSdigital/dis-migration-service/store"
	redirectAPI "github.com/ONSdigital/dis-redirect-api/sdk/go"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	datasetErrors "github.com/ONSdigital/dp-dataset-api/apierrors"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPI "github.com/ONSdigital/dp-dataset-api/sdk"
	datasetAPIMocks "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	filesAPI "github.com/ONSdigital/dp-files-api/sdk"
	filesAPIMocks "github.com/ONSdigital/dp-files-api/sdk/mocks"
	topicAPI "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	uploadSDK "github.com/ONSdigital/dp-upload-service/sdk"
	uploadSDKMocks "github.com/ONSdigital/dp-upload-service/sdk/mocks"
	"github.com/ONSdigital/log.go/v2/log"
)

// ExternalServiceList holds the initialiser and initialisation
// state of external services.
type ExternalServiceList struct {
	HealthCheck bool
	Init        Initialiser
	MongoDB     bool
	Migrator    bool
	SlackClient bool
	TopicCache  bool
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

// GetHealthCheck creates a healthcheck with versionInfo and sets the
// HealthCheck flag to true
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

// GetSlackClient returns a Slack client for sending notifications
func (e *ExternalServiceList) GetSlackClient(ctx context.Context, cfg *config.Config) (slack.Clienter, error) {
	slackClient, err := e.Init.DoGetSlackClient(ctx, cfg)
	if err != nil {
		log.Error(ctx, "failed to initialise slack client", err)
		return nil, err
	}
	e.SlackClient = true
	return slackClient, nil
}

// GetMigrator returns the background migrator
func (e *ExternalServiceList) GetMigrator(ctx context.Context, cfg *config.Config, jobService application.JobService, clientList *clients.ClientList, slackClient slack.Clienter) (migrator.Migrator, error) {
	mig, err := e.Init.DoGetMigrator(ctx, cfg, jobService, clientList, slackClient)
	if err != nil {
		return nil, err
	}

	e.Migrator = true
	return mig, nil
}

// GetAppClients gets the app clients for the service
func (e *ExternalServiceList) GetAppClients(ctx context.Context, cfg *config.Config) *clients.ClientList {
	return e.Init.DoGetAppClients(ctx, cfg)
}

// GetTopicCache returns the topic cache
func (e *ExternalServiceList) GetTopicCache(ctx context.Context, cfg *config.Config, clientList *clients.ClientList) (*cache.TopicCache, error) {
	topicCache, err := e.Init.DoGetTopicCache(ctx, cfg, clientList)
	if err != nil {
		return nil, err
	}
	e.TopicCache = true
	return topicCache, nil
}

// DoGetHTTPServer creates an HTTP Server with the provided
// bind address and router
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
	mongodb := &mongo.Mongo{
		MongoConfig: cfg,
	}
	if err := mongodb.Init(ctx); err != nil {
		return nil, err
	}
	log.Info(ctx, "listening to mongo db session", log.Data{"mongo_uri": mongodb.ClusterEndpoint})
	return mongodb, nil
}

// DoGetSlackClient returns a Slack client based on configuration
func (e *Init) DoGetSlackClient(ctx context.Context, cfg *config.Config) (slack.Clienter, error) {
	slackCfg := &slack.Config{
		Enabled:  cfg.SlackConfig.Enabled,
		APIToken: cfg.SlackConfig.APIToken,
		Channels: slack.Channels{
			InfoChannel:    cfg.SlackConfig.Channels.InfoChannel,
			WarningChannel: cfg.SlackConfig.Channels.WarningChannel,
			AlarmChannel:   cfg.SlackConfig.Channels.AlarmChannel,
		},
	}

	slackClient, err := slack.New(slackCfg)
	if err != nil {
		return nil, err
	}

	if slackCfg.Enabled {
		log.Info(ctx, "slack client initialised and enabled")
	} else {
		log.Info(ctx, "slack client initialised (disabled - using noop client)")
	}

	return slackClient, nil
}

// DoGetMigrator returns a Migrator
func (e *Init) DoGetMigrator(ctx context.Context, cfg *config.Config, jobService application.JobService, clientList *clients.ClientList, slackClient slack.Clienter) (migrator.Migrator, error) {
	mig := migrator.NewDefaultMigrator(cfg, jobService, clientList, slackClient)
	log.Info(ctx, "migrator initialised")
	return mig, nil
}

// DoGetTopicCache creates and initializes the topic cache
func (e *Init) DoGetTopicCache(ctx context.Context, cfg *config.Config, clientList *clients.ClientList) (*cache.TopicCache, error) {
	// Create topic cache with update interval
	topicCache, err := cache.NewTopicCache(ctx, &cfg.TopicCacheUpdateInterval)
	if err != nil {
		log.Error(ctx, "failed to create topic cache", err)
		return nil, err
	}

	// Add update function to populate the cache
	topicCache.AddUpdateFunc(
		topicCache.GetTopicCacheKey(),
		cache.UpdateTopicCache(ctx, cfg.ServiceAuthToken, clientList.TopicAPI),
	)

	// Start the cache updates in the background (this will trigger the initial update and periodic updates)
	// Create a dedicated error channel for cache updates
	cacheErrorChan := make(chan error, 1)
	go func() {
		topicCache.StartUpdates(ctx, cacheErrorChan)
	}()

	// Listen for cache errors and log them
	go func() {
		for err := range cacheErrorChan {
			if err != nil {
				log.Error(ctx, "topic cache update error", err)
			}
		}
	}()

	log.Info(ctx, "topic cache initialised", log.Data{
		"update_interval": cfg.TopicCacheUpdateInterval,
	})
	return topicCache, nil
}

// DoGetAppClients returns a set of app clients for the migration service
func (e *Init) DoGetAppClients(ctx context.Context, cfg *config.Config) *clients.ClientList {
	if cfg.EnableMockClients {
		log.Info(ctx, "returning mock app clients")
		return &clients.ClientList{
			DatasetAPI: &datasetAPIMocks.ClienterMock{
				GetDatasetFunc: func(ctx context.Context, headers datasetAPI.Headers, datasetID string) (datasetModels.Dataset, error) {
					return datasetModels.Dataset{}, datasetErrors.ErrDatasetNotFound
				},
			},
			FilesAPI:      &filesAPIMocks.ClienterMock{},
			RedirectAPI:   &clientMocks.RedirectAPIClientMock{},
			TopicAPI:      nil, // Mock topic API client can be added if needed
			UploadService: &uploadSDKMocks.ClienterMock{},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetPageDataFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedee.PageData, error) {
					return zebedee.PageData{
						Type: "dataset_landing_page",
						Description: zebedee.Description{
							Title: "Mock Dataset Title",
						},
					}, nil
				},
				GetDatasetLandingPageFunc: func(ctx context.Context, userAccessToken, collectionID, lang, path string) (zebedee.DatasetLandingPage, error) {
					return zebedee.DatasetLandingPage{
						Type: "dataset_landing_page",
						Description: zebedee.Description{
							Title: "Mock Dataset Title",
							Contact: zebedee.Contact{
								Name:      "Mock Contact Name",
								Email:     "mock.contact@example.com",
								Telephone: "my telephone",
							},
							Keywords:    []string{"some", "keywords", "here"},
							NextRelease: "2024-12-31",
						},
						RelatedMethodology: []zebedee.Link{
							{
								URI:     "/qmi/real-qmi",
								Title:   "This is a real QMI",
								Summary: "This is the summary of the real QMI",
							},
						},
						Datasets: []zebedee.Link{
							{
								URI: "/mock-dataset/editions/2024",
							},
						},
					}, nil
				},
			},
		}
	}

	log.Info(ctx, "initialising app clients")
	topicAPIClient := topicAPI.New(cfg.TopicAPIURL)

	return &clients.ClientList{
		DatasetAPI:    datasetAPI.New(cfg.DatasetAPIURL),
		FilesAPI:      filesAPI.New(cfg.FilesAPIURL, cfg.ServiceAuthToken),
		RedirectAPI:   redirectAPI.NewClient(cfg.RedirectAPIURL),
		TopicAPI:      topicAPIClient,
		UploadService: uploadSDK.New(cfg.UploadServiceURL),
		Zebedee:       zebedee.New(cfg.ZebedeeURL),
	}
}

// DoGetAuthorisationMiddleware creates authorisation middleware for the given
// config
func (e *Init) DoGetAuthorisationMiddleware(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
	return authorisation.NewFeatureFlaggedMiddleware(ctx, authorisationConfig, nil)
}

// GetAuthorisationMiddleware gets the authorisation middleware for the service
func (e *ExternalServiceList) GetAuthorisationMiddleware(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
	return e.Init.DoGetAuthorisationMiddleware(ctx, authorisationConfig)
}
