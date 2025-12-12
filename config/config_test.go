package config

import (
	"os"
	"testing"
	"time"

	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	dpMongo "github.com/ONSdigital/dp-mongodb/v3/mongodb"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	os.Clearenv()
	var err error
	var configuration *Config

	Convey("Given an environment with no environment variables set", t, func() {
		Convey("Then cfg should be nil", func() {
			So(cfg, ShouldBeNil)
		})

		Convey("When the config values are retrieved", func() {
			Convey("Then there should be no error returned, and values are as expected", func() {
				configuration, err = Get() // This Get() is only called once, when inside this function
				So(err, ShouldBeNil)
				So(configuration, ShouldResemble, &Config{
					BindAddr:                        "localhost:30100",
					DatasetAPIURL:                   "http://localhost:22000",
					DefaultLimit:                    10,
					DefaultOffset:                   0,
					DefaultMaxLimit:                 100,
					EnableMockClients:               false,
					FilesAPIURL:                     "http://localhost:26900",
					GracefulShutdownTimeout:         5 * time.Second,
					HealthCheckInterval:             30 * time.Second,
					HealthCheckCriticalTimeout:      90 * time.Second,
					OTBatchTimeout:                  5 * time.Second,
					OTExporterOTLPEndpoint:          "localhost:4317",
					OTServiceName:                   "dis-migration-service",
					OtelEnabled:                     false,
					MigratorMaxConcurrentExecutions: 5,
					MigratorPollInterval:            5 * time.Second,
					MongoConfig: MongoConfig{
						MongoDriverConfig: dpMongo.MongoDriverConfig{
							ClusterEndpoint:               "localhost:27017",
							Username:                      "",
							Password:                      "",
							Database:                      "migrations",
							Collections:                   map[string]string{CountersCollectionTitle: CountersCollectionName, JobsCollectionTitle: JobsCollectionName, EventsCollectionTitle: EventsCollectionName, TasksCollectionTitle: TasksCollectionName},
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
					AuthConfig:       authorisation.NewDefaultConfig(),
					RedirectAPIURL:   "http://localhost:29900",
					UploadServiceURL: "http://localhost:25100",
					ZebedeeURL:       "http://localhost:8082",
				})
			})

			Convey("Then a second call to config should return the same config", func() {
				// This achieves code coverage of the first return in the Get() function.
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, cfg)
			})
		})
	})
}
