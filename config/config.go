package config

import (
	"time"

	dpMongo "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/kelseyhightower/envconfig"
)

// MongoConfig holds MongoDB configuration settings.
type MongoConfig struct {
	dpMongo.MongoDriverConfig
}

// Config represents service configuration for dis-migration-service
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	EnableMockClients          bool          `envconfig:"ENABLE_MOCK_CLIENTS"`
	FilesAPIURL                string        `envconfig:"FILES_API_URL"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	MigrationServiceURL        string        `envconfig:"MIGRATION_SERVICE_URL"`
	OTBatchTimeout             time.Duration `encconfig:"OTEL_BATCH_TIMEOUT"`
	OTExporterOTLPEndpoint     string        `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTServiceName              string        `envconfig:"OTEL_SERVICE_NAME"`
	OtelEnabled                bool          `envconfig:"OTEL_ENABLED"`
	RedirectAPIURL             string        `envconfig:"REDIRECT_API_URL"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"`
	UploadServiceURL           string        `envconfig:"UPLOAD_SERVICE_URL"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
	MongoConfig
}

var cfg *Config

const (
	// JobsCollectionTitle is the well known name of the MongoDB collection for
	// migration jobs.
	JobsCollectionTitle = "MigrationsJobsCollection"
	// JobsCollectionName is the actual name of the MongoDB collection for
	// migration jobs.
	JobsCollectionName = "jobs"
	// EventsCollectionTitle is the well known name of the MongoDB collection
	// for migration events.
	EventsCollectionTitle = "MigrationsEventsCollection"
	// EventsCollectionName is the actual name of the MongoDB collection for
	// migration events.
	EventsCollectionName = "events"
	// TasksCollectionTitle is the well known name of the MongoDB collection
	// for migration tasks.
	TasksCollectionTitle = "MigrationsTasksCollection"
	// TasksCollectionName is the actual name of the MongoDB collection for
	// migration tasks.
	TasksCollectionName = "tasks"
)

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                   "localhost:30100",
		DatasetAPIURL:              "http://localhost:22000",
		EnableMockClients:          false,
		FilesAPIURL:                "http://localhost:26900",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		OTBatchTimeout:             5 * time.Second,
		OTExporterOTLPEndpoint:     "localhost:4317",
		OTServiceName:              "dis-migration-service",
		OtelEnabled:                false,
		MigrationServiceURL:        "http://localhost:30100",
		MongoConfig: MongoConfig{
			MongoDriverConfig: dpMongo.MongoDriverConfig{
				ClusterEndpoint:               "localhost:27017",
				Username:                      "",
				Password:                      "",
				Database:                      "migrations",
				Collections:                   map[string]string{JobsCollectionTitle: JobsCollectionName, EventsCollectionTitle: EventsCollectionName, TasksCollectionTitle: TasksCollectionName},
				ReplicaSet:                    "",
				IsStrongReadConcernEnabled:    false,
				IsWriteConcernMajorityEnabled: true,
				ConnectTimeout:                5 * time.Second,
				QueryTimeout:                  15 * time.Second,
				TLSConnectionConfig: dpMongo.TLSConnectionConfig{
					IsSSL: false,
				},
			},
		},
		RedirectAPIURL:   "http://localhost:29900",
		UploadServiceURL: "http://localhost:25100",
		ZebedeeURL:       "http://localhost:8082",
	}

	return cfg, envconfig.Process("", cfg)
}
