package mongo_test

import (
	"testing"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/mongo"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMongoActualCollectionName(t *testing.T) {
	Convey("Given a Mongo instance with custom collection names", t, func() {
		mongoInstance := &mongo.Mongo{
			MongoConfig: config.MongoConfig{
				MongoDriverConfig: mongodriver.MongoDriverConfig{
					ClusterEndpoint: "localhost:27017",
					Database:        "test-db",
					Collections: map[string]string{
						config.JobsCollectionTitle:     "prefix_jobs",
						config.TasksCollectionTitle:    "prefix_tasks",
						config.EventsCollectionTitle:   "prefix_events",
						config.CountersCollectionTitle: "prefix_counters",
					},
				},
			},
		}

		Convey("When getting the actual collection name for jobs", func() {
			actualName := mongoInstance.ActualCollectionName(config.JobsCollectionTitle)

			Convey("Then the custom name should be returned", func() {
				So(actualName, ShouldEqual, "prefix_jobs")
			})
		})

		Convey("When getting the actual collection name for tasks", func() {
			actualName := mongoInstance.ActualCollectionName(config.TasksCollectionTitle)

			Convey("Then the custom name should be returned", func() {
				So(actualName, ShouldEqual, "prefix_tasks")
			})
		})

		Convey("When getting the actual collection name for events", func() {
			actualName := mongoInstance.ActualCollectionName(config.EventsCollectionTitle)

			Convey("Then the custom name should be returned", func() {
				So(actualName, ShouldEqual, "prefix_events")
			})
		})

		Convey("When getting the actual collection name for counters", func() {
			actualName := mongoInstance.ActualCollectionName(config.CountersCollectionTitle)

			Convey("Then the custom name should be returned", func() {
				So(actualName, ShouldEqual, "prefix_counters")
			})
		})
	})

	Convey("Given a Mongo instance with default collection names", t, func() {
		mongoInstance := &mongo.Mongo{
			MongoConfig: config.MongoConfig{
				MongoDriverConfig: mongodriver.MongoDriverConfig{
					ClusterEndpoint: "localhost:27017",
					Database:        "test-db",
					Collections: map[string]string{
						config.JobsCollectionTitle:     config.JobsCollectionName,
						config.TasksCollectionTitle:    config.TasksCollectionName,
						config.EventsCollectionTitle:   config.EventsCollectionName,
						config.CountersCollectionTitle: config.CountersCollectionName,
					},
				},
			},
		}

		Convey("When getting the actual collection name", func() {
			actualName := mongoInstance.ActualCollectionName(config.JobsCollectionTitle)

			Convey("Then the default name should be returned", func() {
				So(actualName, ShouldEqual, "jobs")
			})
		})
	})
}

func TestCollectionValidation(t *testing.T) {
	Convey("Given a Mongo configuration", t, func() {
		Convey("When validating collection names", func() {
			mongoInstance := &mongo.Mongo{
				MongoConfig: config.MongoConfig{
					MongoDriverConfig: mongodriver.MongoDriverConfig{
						ClusterEndpoint: "localhost:27017",
						Database:        "test-db",
						Collections: map[string]string{
							config.JobsCollectionTitle:     "test-jobs",
							config.TasksCollectionTitle:    "test-tasks",
							config.EventsCollectionTitle:   "test-events",
							config.CountersCollectionTitle: "test-counters",
						},
					},
				},
			}

			Convey("All expected collections should have correct names", func() {
				jobsCollection := mongoInstance.ActualCollectionName(config.JobsCollectionTitle)
				tasksCollection := mongoInstance.ActualCollectionName(config.TasksCollectionTitle)
				eventsCollection := mongoInstance.ActualCollectionName(config.EventsCollectionTitle)
				countersCollection := mongoInstance.ActualCollectionName(config.CountersCollectionTitle)

				So(jobsCollection, ShouldEqual, "test-jobs")
				So(tasksCollection, ShouldEqual, "test-tasks")
				So(eventsCollection, ShouldEqual, "test-events")
				So(countersCollection, ShouldEqual, "test-counters")
			})
		})
	})
}
