package mongo_test

import (
	"context"
	"testing"
	"time"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/mongo"
	mongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
)

type TaskList []*domain.Task

func (tl TaskList) AsInterfaceList() []interface{} {
	result := make([]interface{}, len(tl))
	for i, task := range tl {
		result[i] = task
	}
	return result
}

func setupTaskStoreTest(t *testing.T, ctx context.Context) (*mongo.Mongo, *mongoDriver.MongoConnection) {
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

func TestCreateTask(t *testing.T) {
	Convey("Given a MongoDB connection and tasks collection", t, func() {
		ctx := context.Background()
		m, conn := setupTaskStoreTest(t, ctx)
		collection := config.TasksCollectionName

		Convey("When creating a new task", func() {
			task := &domain.Task{
				ID:          "task-1",
				JobID:       "job-123",
				State:       domain.JobStateSubmitted,
				LastUpdated: time.Now().UTC(),
			}

			err := m.CreateTask(ctx, task)

			Convey("Then the task should be created without error", func() {
				So(err, ShouldBeNil)
				var retrieved domain.Task
				err := queryMongo(conn, collection, bson.M{"_id": task.ID}, &retrieved)
				So(err, ShouldBeNil)
				So(retrieved.ID, ShouldEqual, task.ID)
				So(retrieved.JobID, ShouldEqual, "job-123")
				So(retrieved.State, ShouldEqual, domain.JobStateSubmitted)
			})

			Reset(func() {
				conn.DropDatabase(ctx)
			})
		})

		Convey("When creating multiple tasks", func() {
			task1 := &domain.Task{
				ID:          "task-1",
				JobID:       "job-123",
				State:       domain.JobStateSubmitted,
				LastUpdated: time.Now().UTC(),
			}
			task2 := &domain.Task{
				ID:          "task-2",
				JobID:       "job-123",
				State:       domain.JobStateApproved,
				LastUpdated: time.Now().UTC(),
			}

			err1 := m.CreateTask(ctx, task1)
			err2 := m.CreateTask(ctx, task2)

			Convey("Then both tasks should be created without error with unique IDs", func() {
				So(err1, ShouldBeNil)
				So(err2, ShouldBeNil)
				So(task1.ID, ShouldNotEqual, task2.ID)
			})

			Reset(func() {
				conn.DropDatabase(ctx)
			})
		})
	})
}

func TestGetJobTasks(t *testing.T) {
	Convey("Given a MongoDB connection with tasks for multiple jobs", t, func() {
		ctx := context.Background()
		m, conn := setupTaskStoreTest(t, ctx)
		collection := config.TasksCollectionName

		now := time.Now().UTC().Truncate(time.Millisecond)
		jobID := "job-789"
		otherJobID := "job-other"

		task1 := &domain.Task{
			ID:          "task-1",
			JobID:       jobID,
			State:       domain.JobStateSubmitted,
			LastUpdated: now.Add(-3 * time.Hour),
		}
		task2 := &domain.Task{
			ID:          "task-2",
			JobID:       jobID,
			State:       domain.JobStateApproved,
			LastUpdated: now.Add(-2 * time.Hour),
		}
		task3 := &domain.Task{
			ID:          "task-3",
			JobID:       jobID,
			State:       domain.JobStateCompleted,
			LastUpdated: now.Add(-1 * time.Hour),
		}
		task4 := &domain.Task{
			ID:          "task-4",
			JobID:       jobID,
			State:       domain.JobStateMigrating,
			LastUpdated: now,
		}
		taskOtherJob := &domain.Task{
			ID:          "task-other",
			JobID:       otherJobID,
			State:       domain.JobStateSubmitted,
			LastUpdated: now,
		}

		testData := TaskList{task1, task2, task3, task4, taskOtherJob}

		if err := setUpTestDataTasks(ctx, conn, collection, testData); err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}

		Convey("When retrieving all tasks for a job", func() {
			retrieved, count, err := m.GetJobTasks(ctx, jobID, 10, 0)

			Convey("Then the operation should succeed with correct data and sorting", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 4)
				So(len(retrieved), ShouldEqual, 4)
				for _, task := range retrieved {
					So(task.JobID, ShouldEqual, jobID)
					So(task.ID, ShouldNotEqual, "task-other")
				}
				So(retrieved[0].ID, ShouldEqual, task4.ID) // Most recent
				So(retrieved[3].ID, ShouldEqual, task1.ID) // Oldest
			})
		})

		Convey("When retrieving tasks with limit and offset", func() {
			page1, count1, err1 := m.GetJobTasks(ctx, jobID, 2, 0)
			page2, count2, err2 := m.GetJobTasks(ctx, jobID, 10, 2)

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

		Convey("When retrieving tasks for a non-existent job", func() {
			retrieved, count, err := m.GetJobTasks(ctx, "non-existent-job", 10, 0)

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

func TestCountTasksByJobID(t *testing.T) {
	Convey("Given a MongoDB connection with tasks for multiple jobs", t, func() {
		ctx := context.Background()
		m, conn := setupTaskStoreTest(t, ctx)
		collection := config.TasksCollectionName

		now := time.Now().UTC().Truncate(time.Millisecond)
		jobID := "job-count-test"
		otherJobID := "job-other-count"

		task1 := &domain.Task{
			ID:          "task-1",
			JobID:       jobID,
			State:       domain.JobStateSubmitted,
			LastUpdated: now,
		}
		task2 := &domain.Task{
			ID:          "task-2",
			JobID:       jobID,
			State:       domain.JobStateApproved,
			LastUpdated: now,
		}
		task3 := &domain.Task{
			ID:          "task-3",
			JobID:       jobID,
			State:       domain.JobStateCompleted,
			LastUpdated: now,
		}
		taskOtherJob := &domain.Task{
			ID:          "task-other-count",
			JobID:       otherJobID,
			State:       domain.JobStateSubmitted,
			LastUpdated: now,
		}

		testData := TaskList{task1, task2, task3, taskOtherJob}

		if err := setUpTestDataTasks(ctx, conn, collection, testData); err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}

		Convey("When counting tasks for different jobs", func() {
			count1, err1 := m.CountTasksByJobID(ctx, jobID)
			count2, err2 := m.CountTasksByJobID(ctx, otherJobID)
			count3, err3 := m.CountTasksByJobID(ctx, "non-existent-job")

			Convey("Then counts should be correct for each job", func() {
				So(err1, ShouldBeNil)
				So(err2, ShouldBeNil)
				So(err3, ShouldBeNil)
				So(count1, ShouldEqual, 3)
				So(count2, ShouldEqual, 1)
				So(count3, ShouldEqual, 0)
				So(count1, ShouldNotEqual, count2)
			})
		})

		Convey("When counting tasks for the same job multiple times", func() {
			count1, err1 := m.CountTasksByJobID(ctx, jobID)
			count2, err2 := m.CountTasksByJobID(ctx, jobID)

			Convey("Then counts should be consistent", func() {
				So(err1, ShouldBeNil)
				So(err2, ShouldBeNil)
				So(count1, ShouldEqual, count2)
				So(count1, ShouldEqual, 3)
			})
		})

		Reset(func() {
			conn.DropDatabase(ctx)
		})
	})
}

func setUpTestDataTasks(ctx context.Context, mongoConnection *mongoDriver.MongoConnection, collection string, tasks TaskList) error {
	if err := mongoConnection.DropDatabase(ctx); err != nil {
		return err
	}

	if _, err := mongoConnection.Collection(collection).InsertMany(
		ctx,
		tasks.AsInterfaceList(),
	); err != nil {
		return err
	}

	return nil
}
