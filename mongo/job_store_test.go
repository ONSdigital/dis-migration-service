package mongo_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dis-migration-service/mongo"
	mongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	. "github.com/smartystreets/goconvey/convey"
	testMongoContainer "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

type JobList []*domain.Job

func (jl JobList) AsInterfaceList() []interface{} {
	result := make([]interface{}, len(jl))
	for i, job := range jl {
		result[i] = job
	}
	return result
}

func setupJobStoreTest(t *testing.T, _ context.Context) (*mongo.Mongo, *mongoDriver.MongoConnection) {
	conn, err := setupSharedMongo(t)
	if err != nil {
		t.Fatalf("failed to setup shared mongo: %v", err)
	}

	m := &mongo.Mongo{
		MongoConfig: config.MongoConfig{
			MongoDriverConfig: mongoDriver.MongoDriverConfig{
				Database: database,
				Collections: map[string]string{
					config.JobsCollectionTitle:   config.JobsCollectionName,
					config.EventsCollectionTitle: config.EventsCollectionName,
					config.TasksCollectionTitle:  config.TasksCollectionName,
				},
			},
		},
		Connection: conn,
	}

	return m, conn
}
func TestCreateJob(t *testing.T) {
	Convey("Given a MongoDB connection and jobs collection", t, func() {
		ctx := context.Background()
		m, conn := setupJobStoreTest(t, ctx)
		collection := config.JobsCollectionName

		Convey("When creating a new job", func() {
			jobConfig := &domain.JobConfig{
				SourceID: "source-1",
				TargetID: "target-1",
				Type:     "migration",
			}
			job := domain.NewJob(jobConfig, 1, "Test Dataset Title")

			err := m.CreateJob(ctx, &job)

			Convey("Then the job should be created without error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the job should be retrievable from the database", func() {
				var retrieved domain.Job
				err := queryMongo(conn, collection, bson.M{"_id": job.ID}, &retrieved)
				So(err, ShouldBeNil)
				So(retrieved.ID, ShouldEqual, job.ID)
				So(retrieved.State, ShouldEqual, domain.StateSubmitted)
				So(retrieved.Config.SourceID, ShouldEqual, jobConfig.SourceID)
			})

			Reset(func() {
				conn.DropDatabase(ctx)
			})
		})

		Convey("When creating multiple jobs", func() {
			jobConfig1 := &domain.JobConfig{
				SourceID: "source-1",
				TargetID: "target-1",
				Type:     "migration",
			}
			jobConfig2 := &domain.JobConfig{
				SourceID: "source-2",
				TargetID: "target-2",
				Type:     "sync",
			}

			job1 := domain.NewJob(jobConfig1, 1, "Test Dataset Title 1")
			job2 := domain.NewJob(jobConfig2, 1, "Test Dataset Title 2")

			err1 := m.CreateJob(ctx, &job1)
			err2 := m.CreateJob(ctx, &job2)

			Convey("Then both jobs should be created without error", func() {
				So(err1, ShouldBeNil)
				So(err2, ShouldBeNil)
				So(job1.ID, ShouldNotEqual, job2.ID)
			})

			Reset(func() {
				conn.DropDatabase(ctx)
			})
		})
	})
}

func TestGetJob(t *testing.T) {
	Convey("Given a MongoDB connection with a stored job", t, func() {
		ctx := context.Background()
		m, conn := setupJobStoreTest(t, ctx)
		collection := config.JobsCollectionName

		jobConfig := &domain.JobConfig{
			SourceID: "source-retrieve",
			TargetID: "target-retrieve",
			Type:     "migration",
		}
		testJob := domain.NewJob(jobConfig, 1, "Test Dataset Title")

		if err := setUpTestData(ctx, conn, collection, JobList{&testJob}); err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}

		Convey("When retrieving an existing job by ID", func() {
			retrieved, err := m.GetJob(ctx, testJob.JobNumber)

			Convey("Then the operation should succeed without error", func() {
				So(err, ShouldBeNil)
				So(retrieved, ShouldNotBeNil)
				So(retrieved.ID, ShouldEqual, testJob.ID)
				So(retrieved.State, ShouldEqual, testJob.State)
				So(retrieved.Config.SourceID, ShouldEqual, testJob.Config.SourceID)
			})
		})

		Convey("When retrieving a non-existent job", func() {
			retrieved, err := m.GetJob(ctx, 99999)

			Convey("Then an error should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrJobNotFound)
				So(retrieved, ShouldBeNil)
			})
		})

		Reset(func() {
			conn.DropDatabase(ctx)
		})
	})
}

func TestGetJobs(t *testing.T) {
	Convey("Given a MongoDB connection with multiple jobs in different states", t, func() {
		ctx := context.Background()
		m, conn := setupJobStoreTest(t, ctx)
		collection := config.JobsCollectionName

		now := time.Now().UTC().Truncate(time.Millisecond)

		config1 := &domain.JobConfig{SourceID: "s1", TargetID: "t1", Type: "migration"}
		config2 := &domain.JobConfig{SourceID: "s2", TargetID: "t2", Type: "sync"}
		config3 := &domain.JobConfig{SourceID: "s3", TargetID: "t3", Type: "migration"}
		config4 := &domain.JobConfig{SourceID: "s4", TargetID: "t4", Type: "sync"}

		job1 := domain.NewJob(config1, 1, "Test Dataset Title")
		job1.State = domain.StateSubmitted
		job1.LastUpdated = now.Add(-3 * time.Hour)

		job2 := domain.NewJob(config2, 1, "Test Dataset Title")
		job2.State = domain.StateSubmitted
		job2.LastUpdated = now.Add(-2 * time.Hour)

		job3 := domain.NewJob(config3, 1, "Test Dataset Title")
		job3.State = domain.StateApproved
		job3.LastUpdated = now.Add(-1 * time.Hour)

		job4 := domain.NewJob(config4, 1, "Test Dataset Title")
		job4.State = domain.StateApproved
		job4.LastUpdated = now

		testData := JobList{&job1, &job2, &job3, &job4}

		if err := setUpTestData(ctx, conn, collection, testData); err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}

		Convey("When retrieving all jobs without state filter", func() {
			retrieved, count, err := m.GetJobs(ctx, []domain.State{}, 10, 0)

			Convey("Then the operation should succeed and return all jobs sorted by last_updated", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 4)
				So(len(retrieved), ShouldEqual, 4)
				So(retrieved[0].ID, ShouldEqual, job4.ID) // Most recent
				So(retrieved[3].ID, ShouldEqual, job1.ID) // Oldest
			})
		})

		Convey("When retrieving jobs with a single state filter", func() {
			stateFilter := []domain.State{domain.StateSubmitted}
			retrieved, count, err := m.GetJobs(ctx, stateFilter, 10, 0)

			Convey("Then only jobs with the specified state should be returned", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 2)
				So(len(retrieved), ShouldEqual, 2)
				for _, job := range retrieved {
					So(job.State, ShouldEqual, domain.StateSubmitted)
				}
			})
		})

		Convey("When retrieving jobs with multiple state filters", func() {
			stateFilter := []domain.State{
				domain.StateSubmitted,
				domain.StateApproved,
			}
			retrieved, count, err := m.GetJobs(ctx, stateFilter, 10, 0)

			Convey("Then all jobs matching any state should be returned", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 4)
				So(len(retrieved), ShouldEqual, 4)
			})
		})

		Convey("When retrieving jobs with limit and offset", func() {
			page1, count1, err1 := m.GetJobs(ctx, []domain.State{}, 2, 0)
			page2, count2, err2 := m.GetJobs(ctx, []domain.State{}, 2, 2)

			Convey("Then pagination should work correctly", func() {
				So(err1, ShouldBeNil)
				So(err2, ShouldBeNil)
				So(count1, ShouldEqual, 4)
				So(count2, ShouldEqual, 4)
				So(len(page1), ShouldEqual, 2)
				So(len(page2), ShouldEqual, 2)
				So(page1[0].ID, ShouldNotEqual, page2[0].ID)
			})
		})

		Convey("When retrieving jobs with a filter that matches nothing", func() {
			stateFilter := []domain.State{domain.StateCancelled}
			retrieved, count, err := m.GetJobs(ctx, stateFilter, 10, 0)

			Convey("Then an empty list should be returned", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 0)
				So(len(retrieved), ShouldEqual, 0)
			})
		})

		Reset(func() {
			conn.DropDatabase(ctx)
		})
	})
}

func TestGetJobsByConfigAndState(t *testing.T) {
	Convey("Given a MongoDB connection with jobs from multiple configs", t, func() {
		ctx := context.Background()
		m, conn := setupJobStoreTest(t, ctx)
		collection := config.JobsCollectionName

		now := time.Now().UTC().Truncate(time.Millisecond)

		config1 := &domain.JobConfig{
			SourceID: "source-1",
			TargetID: "target-1",
			Type:     "migration",
		}
		config2 := &domain.JobConfig{
			SourceID: "source-2",
			TargetID: "target-2",
			Type:     "sync",
		}

		job1 := domain.NewJob(config1, 1, "Test Dataset Title")
		job1.State = domain.StateSubmitted
		job1.LastUpdated = now

		job2 := domain.NewJob(config1, 1, "Test Dataset Title")
		job2.State = domain.StateApproved
		job2.LastUpdated = now

		job3 := domain.NewJob(config1, 1, "Test Dataset Title")
		job3.State = domain.StateCompleted
		job3.LastUpdated = now

		job4 := domain.NewJob(config2, 1, "Test Dataset Title")
		job4.State = domain.StateSubmitted
		job4.LastUpdated = now

		testData := JobList{&job1, &job2, &job3, &job4}

		if err := setUpTestData(ctx, conn, collection, testData); err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}

		Convey("When retrieving jobs by config and single state", func() {
			stateFilter := []domain.State{domain.StateSubmitted}
			retrieved, err := m.GetJobsByConfigAndState(ctx, config1, stateFilter, 10, 0)

			Convey("Then only jobs matching config and state should be returned", func() {
				So(err, ShouldBeNil)
				So(len(retrieved), ShouldEqual, 1)
				So(retrieved[0].ID, ShouldEqual, job1.ID)
				So(retrieved[0].State, ShouldEqual, domain.StateSubmitted)
				So(retrieved[0].Config.SourceID, ShouldEqual, config1.SourceID)
			})
		})

		Convey("When retrieving jobs by config with multiple states or empty filter", func() {
			// Multiple states
			retrieved1, err1 := m.GetJobsByConfigAndState(ctx, config1, []domain.State{
				domain.StateSubmitted,
				domain.StateApproved,
			}, 10, 0)

			// Empty filter
			retrieved2, err2 := m.GetJobsByConfigAndState(ctx, config1, []domain.State{}, 10, 0)

			Convey("Then all matching states should be returned", func() {
				So(err1, ShouldBeNil)
				So(len(retrieved1), ShouldEqual, 2)

				So(err2, ShouldBeNil)
				So(len(retrieved2), ShouldEqual, 3)
				for _, job := range retrieved2 {
					So(job.Config.SourceID, ShouldEqual, config1.SourceID)
				}
			})
		})

		Convey("When retrieving jobs by non-matching config or with pagination", func() {
			nonExistentConfig := &domain.JobConfig{
				SourceID: "non-existent",
				TargetID: "non-existent",
				Type:     "migration",
			}
			retrieved1, err1 := m.GetJobsByConfigAndState(ctx, nonExistentConfig, []domain.State{domain.StateSubmitted}, 10, 0)

			retrieved2, err2 := m.GetJobsByConfigAndState(ctx, config1, []domain.State{}, 1, 0)
			retrieved3, err3 := m.GetJobsByConfigAndState(ctx, config1, []domain.State{}, 10, 2)

			Convey("Then correct filtering and pagination should apply", func() {
				So(err1, ShouldBeNil)
				So(len(retrieved1), ShouldEqual, 0)

				So(err2, ShouldBeNil)
				So(len(retrieved2), ShouldEqual, 1)

				So(err3, ShouldBeNil)
				So(len(retrieved3), ShouldEqual, 1)
			})
		})

		Reset(func() {
			conn.DropDatabase(ctx)
		})
	})
}

func TestGetJobsByConfig(t *testing.T) {
	Convey("Given a MongoDB connection with jobs in multiple states for a config", t, func() {
		ctx := context.Background()
		m, conn := setupJobStoreTest(t, ctx)
		collection := config.JobsCollectionName

		now := time.Now().UTC().Truncate(time.Millisecond)

		jobConfig := &domain.JobConfig{
			SourceID: "source-1",
			TargetID: "target-1",
			Type:     "migration",
		}

		job1 := domain.NewJob(jobConfig, 1, "Test Dataset Title")
		job1.State = domain.StateSubmitted
		job1.LastUpdated = now

		job2 := domain.NewJob(jobConfig, 1, "Test Dataset Title")
		job2.State = domain.StateApproved
		job2.LastUpdated = now

		job3 := domain.NewJob(jobConfig, 1, "Test Dataset Title")
		job3.State = domain.StateCompleted
		job3.LastUpdated = now

		testData := JobList{&job1, &job2, &job3}

		if err := setUpTestData(ctx, conn, collection, testData); err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}

		Convey("When retrieving all jobs by config", func() {
			retrieved, err := m.GetJobsByConfig(ctx, jobConfig, 10, 0)

			Convey("Then all jobs in any state should be returned", func() {
				So(err, ShouldBeNil)
				So(len(retrieved), ShouldEqual, 3)
				for _, job := range retrieved {
					So(job.Config.SourceID, ShouldEqual, jobConfig.SourceID)
				}

				states := make(map[domain.State]bool)
				for _, job := range retrieved {
					states[job.State] = true
				}
				So(len(states), ShouldEqual, 3)
			})
		})

		Convey("When retrieving jobs with limit and offset", func() {
			retrieved1, err1 := m.GetJobsByConfig(ctx, jobConfig, 2, 0)
			retrieved2, err2 := m.GetJobsByConfig(ctx, jobConfig, 10, 1)

			Convey("Then pagination should work correctly", func() {
				So(err1, ShouldBeNil)
				So(len(retrieved1), ShouldEqual, 2)

				So(err2, ShouldBeNil)
				So(len(retrieved2), ShouldEqual, 2)
			})
		})

		Convey("When retrieving jobs by non-matching config", func() {
			nonExistentConfig := &domain.JobConfig{
				SourceID: "non-existent",
				TargetID: "non-existent",
				Type:     "sync",
			}
			retrieved, err := m.GetJobsByConfig(ctx, nonExistentConfig, 10, 0)

			Convey("Then an empty list should be returned", func() {
				So(err, ShouldBeNil)
				So(len(retrieved), ShouldEqual, 0)
			})
		})

		Reset(func() {
			conn.DropDatabase(ctx)
		})
	})
}

func TestClaimJob(t *testing.T) {
	Convey("Given a MongoDB connection with jobs in different states", t, func() {
		ctx := context.Background()
		m, conn := setupJobStoreTest(t, ctx)
		collection := config.JobsCollectionName

		jobConfig := &domain.JobConfig{
			SourceID: "source-claim",
			TargetID: "target-claim",
			Type:     "migration",
		}

		// Create jobs with different states and timestamps
		job1 := domain.NewJob(jobConfig, 1, "Job 1")
		job1.State = domain.StateSubmitted
		job1.LastUpdated = time.Now().Add(-10 * time.Minute)

		job2 := domain.NewJob(jobConfig, 2, "Job 2")
		job2.State = domain.StateSubmitted
		job2.LastUpdated = time.Now().Add(-5 * time.Minute)

		job3 := domain.NewJob(jobConfig, 3, "Job 3")
		job3.State = domain.StateInReview
		job3.LastUpdated = time.Now().Add(-2 * time.Minute)

		testData := JobList{&job1, &job2, &job3}
		setUpTestData(ctx, conn, collection, testData)

		Convey("When claiming a submitted job", func() {
			claimed, err := m.ClaimJob(ctx, domain.StateSubmitted, domain.StateInReview)

			Convey("Then the oldest submitted job should be claimed and updated", func() {
				So(err, ShouldBeNil)
				So(claimed, ShouldNotBeNil)
				So(claimed.ID, ShouldEqual, job1.ID)
				So(claimed.State, ShouldEqual, domain.StateInReview)
			})

			Convey("And the job state should be updated in the database", func() {
				var retrieved domain.Job
				queryMongo(conn, collection, bson.M{"_id": claimed.ID}, &retrieved)
				So(retrieved.State, ShouldEqual, domain.StateInReview)
			})

			Convey("And claiming again should return the next oldest job", func() {
				claimed2, err2 := m.ClaimJob(ctx, domain.StateSubmitted, domain.StateInReview)
				So(err2, ShouldBeNil)
				So(claimed2, ShouldNotBeNil)
				So(claimed2.ID, ShouldEqual, job2.ID)
			})
		})

		Convey("When no jobs match the pending state", func() {
			// Claim all submitted jobs first
			m.ClaimJob(ctx, domain.StateSubmitted, domain.StateInReview)
			m.ClaimJob(ctx, domain.StateSubmitted, domain.StateInReview)

			claimed, err := m.ClaimJob(ctx, domain.StateSubmitted, domain.StateInReview)

			Convey("Then nil should be returned without error", func() {
				So(err, ShouldBeNil)
				So(claimed, ShouldBeNil)
			})
		})

		Reset(func() {
			conn.DropDatabase(ctx)
		})
	})
}

func TestGetJobsByState(t *testing.T) {
	Convey("Given a MongoDB connection with jobs in different states", t, func() {
		ctx := context.Background()
		m, conn := setupJobStoreTest(t, ctx)
		collection := config.JobsCollectionName

		jobConfig1 := &domain.JobConfig{
			SourceID: "source-1",
			TargetID: "target-1",
			Type:     "migration",
		}

		job1 := domain.NewJob(jobConfig1, 1, "Job 1")
		job1.State = domain.StateSubmitted

		job2 := domain.NewJob(jobConfig1, 2, "Job 2")
		job2.State = domain.StateInReview

		job3 := domain.NewJob(jobConfig1, 3, "Job 3")
		job3.State = domain.StateCompleted

		job4 := domain.NewJob(jobConfig1, 4, "Job 4")
		job4.State = domain.StateSubmitted

		testData := JobList{&job1, &job2, &job3, &job4}
		setUpTestData(ctx, conn, collection, testData)

		Convey("When retrieving jobs by single state", func() {
			retrieved, count, err := m.GetJobsByState(ctx, []domain.State{domain.StateSubmitted}, 10, 0)

			Convey("Then only jobs matching that state should be returned", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 2)
				So(len(retrieved), ShouldEqual, 2)
				for _, job := range retrieved {
					So(job.State, ShouldEqual, domain.StateSubmitted)
				}
			})
		})

		Convey("When retrieving jobs by multiple states", func() {
			states := []domain.State{domain.StateSubmitted, domain.StateInReview}
			retrieved, count, err := m.GetJobsByState(ctx, states, 10, 0)

			Convey("Then jobs matching any of the states should be returned", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 3)
				So(len(retrieved), ShouldEqual, 3)
			})
		})

		Convey("When using pagination", func() {
			page1, count1, _ := m.GetJobsByState(ctx, []domain.State{domain.StateSubmitted}, 1, 0)
			page2, count2, _ := m.GetJobsByState(ctx, []domain.State{domain.StateSubmitted}, 1, 1)

			Convey("Then pagination should work correctly", func() {
				So(count1, ShouldEqual, 2)
				So(len(page1), ShouldEqual, 1)
				So(count2, ShouldEqual, 2)
				So(len(page2), ShouldEqual, 1)
				So(page1[0].ID, ShouldNotEqual, page2[0].ID)
			})
		})

		Reset(func() {
			conn.DropDatabase(ctx)
		})
	})
}

func TestUpdateJob(t *testing.T) {
	Convey("Given a MongoDB connection with an existing job", t, func() {
		ctx := context.Background()
		m, conn := setupJobStoreTest(t, ctx)
		collection := config.JobsCollectionName

		jobConfig := &domain.JobConfig{
			SourceID: "source-update",
			TargetID: "target-update",
			Type:     "migration",
		}

		job := domain.NewJob(jobConfig, 1, "Original Job")
		job.State = domain.StateSubmitted
		conn.Collection(collection).InsertOne(ctx, job)

		Convey("When updating the job", func() {
			job.State = domain.StateCompleted
			job.Label = "Updated Job"
			err := m.UpdateJob(ctx, &job)

			Convey("Then the job should be updated without error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the changes should be persisted in the database", func() {
				var retrieved domain.Job
				queryMongo(conn, collection, bson.M{"_id": job.ID}, &retrieved)
				So(retrieved.State, ShouldEqual, domain.StateCompleted)
				So(retrieved.Label, ShouldEqual, "Updated Job")
			})
		})

		Convey("When updating a non-existent job", func() {
			nonExistent := domain.NewJob(jobConfig, 999, "Non-existent")
			err := m.UpdateJob(ctx, &nonExistent)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Reset(func() {
			conn.DropDatabase(ctx)
		})
	})
}

func TestUpdateJobState(t *testing.T) {
	Convey("Given a MongoDB connection with an existing job", t, func() {
		ctx := context.Background()
		m, conn := setupJobStoreTest(t, ctx)
		collection := config.JobsCollectionName

		jobConfig := &domain.JobConfig{
			SourceID: "source-state",
			TargetID: "target-state",
			Type:     "migration",
		}

		job := domain.NewJob(jobConfig, 1, "Job for state update")
		job.State = domain.StateSubmitted
		originalTime := job.LastUpdated
		conn.Collection(collection).InsertOne(ctx, job)

		Convey("When updating the job state", func() {
			newTime := time.Now().Add(1 * time.Hour)
			err := m.UpdateJobState(ctx, job.ID, domain.StateInReview, newTime)

			Convey("Then the job state should be updated without error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the state and timestamp should be updated in the database", func() {
				var retrieved domain.Job
				queryMongo(conn, collection, bson.M{"_id": job.ID}, &retrieved)
				So(retrieved.State, ShouldEqual, domain.StateInReview)
				So(retrieved.LastUpdated, ShouldHappenAfter, originalTime)
			})
		})

		Convey("When updating state for a non-existent job", func() {
			err := m.UpdateJobState(ctx, "non-existent-id", domain.StateCompleted, time.Now())

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Reset(func() {
			conn.DropDatabase(ctx)
		})
	})
}

func getMongoDriverConfig(mongoServer *testMongoContainer.MongoDBContainer, database string, collections map[string]string) *mongoDriver.MongoDriverConfig {
	connectionString, err := mongoServer.ConnectionString(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to get mongo server connection string: %v", err))
	}

	connStringURL, err := url.Parse(connectionString)
	if err != nil {
		panic(fmt.Sprintf("failed to parse mongo server connection string: %v", err))
	}

	return &mongoDriver.MongoDriverConfig{
		ConnectTimeout:  5 * time.Second,
		QueryTimeout:    5 * time.Second,
		ClusterEndpoint: connStringURL.Host,
		Database:        database,
		Collections:     collections,
	}
}

func setUpTestData(ctx context.Context, mongoConnection *mongoDriver.MongoConnection, collection string, jobs JobList) error {
	if err := mongoConnection.DropDatabase(ctx); err != nil {
		return err
	}

	if _, err := mongoConnection.Collection(collection).InsertMany(
		ctx,
		jobs.AsInterfaceList(),
	); err != nil {
		return err
	}

	return nil
}

func queryMongo(mongoConnection *mongoDriver.MongoConnection, collection string, query bson.M, res interface{}) error {
	return mongoConnection.Collection(collection).FindOne(context.Background(), query, res)
}
