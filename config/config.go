package config

import (
	"time"

	dpMongo "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/kelseyhightower/envconfig"
)

type MongoConfig struct {
	dpMongo.MongoDriverConfig
}

// Config represents service configuration for dis-migration-service
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	OTBatchTimeout             time.Duration `encconfig:"OTEL_BATCH_TIMEOUT"`
	OTExporterOTLPEndpoint     string        `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTServiceName              string        `envconfig:"OTEL_SERVICE_NAME"`
	OtelEnabled                bool          `envconfig:"OTEL_ENABLED"`
	MongoConfig
}

var cfg *Config

const (
	JobsCollectionTitle   = "MigrationsJobsCollection"
	JobsCollectionName    = "jobs"
	EventsCollectionTitle = "MigrationsEventsCollection"
	EventsCollectionName  = "events"
	TasksCollectionTitle  = "MigrationsTasksCollection"
	TasksCollectionName   = "tasks"
)

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                   "localhost:30100",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		OTBatchTimeout:             5 * time.Second,
		OTExporterOTLPEndpoint:     "localhost:4317",
		OTServiceName:              "dis-migration-service",
		OtelEnabled:                false,
		MongoConfig: MongoConfig{
			MongoDriverConfig: dpMongo.MongoDriverConfig{
				ClusterEndpoint:               "localhost:27017",
				Username:                      "",
				Password:                      "",
				Database:                      "bundles",
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
	}

	return cfg, envconfig.Process("", cfg)
}
