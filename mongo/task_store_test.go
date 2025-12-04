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
				JobNumber:   123,
				State:       domain.StateSubmitted,
				LastUpdated: time.Now().UTC(),
			}

			err := m.CreateTask(ctx, task)

			Convey("Then the task should be created without error", func() {
				So(err, ShouldBeNil)
				var retrieved domain.Task
				err := queryMongo(conn, collection, bson.M{"_id": task.ID}, &retrieved)
				So(err, ShouldBeNil)
				So(retrieved.ID, ShouldEqual, task.ID)
				So(retrieved.JobNumber, ShouldEqual, 123)
				So(retrieved.State, ShouldEqual, domain.StateSubmitted)
			})

			Reset(func() {
				conn.DropDatabase(ctx)
			})
		})

		Convey("When creating multiple tasks", func() {
			task1 := &domain.Task{
				ID:          "task-1",
				JobNumber:   123,
				State:       domain.StateSubmitted,
				LastUpdated: time.Now().UTC(),
			}
			task2 := &domain.Task{
				ID:          "task-2",
				JobNumber:   123,
				State:       domain.StateApproved,
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
		jobID := 789
		otherJobID := 999

		task1 := &domain.Task{
			ID:          "task-1",
			JobNumber:   jobID,
			State:       domain.StateSubmitted,
			LastUpdated: now.Add(-3 * time.Hour),
		}
		task2 := &domain.Task{
			ID:          "task-2",
			JobNumber:   jobID,
			State:       domain.StateApproved,
			LastUpdated: now.Add(-2 * time.Hour),
		}
		task3 := &domain.Task{
			ID:          "task-3",
			JobNumber:   jobID,
			State:       domain.StateCompleted,
			LastUpdated: now.Add(-1 * time.Hour),
		}
		task4 := &domain.Task{
			ID:          "task-4",
			JobNumber:   jobID,
			State:       domain.StateMigrating,
			LastUpdated: now,
		}
		taskOtherJob := &domain.Task{
			ID:          "task-other",
			JobNumber:   otherJobID,
			State:       domain.StateSubmitted,
			LastUpdated: now,
		}

		testData := TaskList{task1, task2, task3, task4, taskOtherJob}

		if err := setUpTestDataTasks(ctx, conn, collection, testData); err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}

		Convey("When retrieving all tasks for a job", func() {
			retrieved, count, err := m.GetJobTasks(ctx, []domain.State{}, jobID, 10, 0)

			Convey("Then the operation should succeed with correct data and sorting", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 4)
				So(len(retrieved), ShouldEqual, 4)
				for _, task := range retrieved {
					So(task.JobNumber, ShouldEqual, jobID)
					So(task.ID, ShouldNotEqual, "task-other")
				}
				So(retrieved[0].ID, ShouldEqual, task4.ID) // Most recent
				So(retrieved[3].ID, ShouldEqual, task1.ID) // Oldest
			})
		})

		Convey("When retrieving tasks with limit and offset", func() {
			page1, count1, err1 := m.GetJobTasks(ctx, []domain.State{}, jobID, 2, 0)
			page2, count2, err2 := m.GetJobTasks(ctx, []domain.State{}, jobID, 10, 2)

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
			retrieved, count, err := m.GetJobTasks(ctx, []domain.State{}, 99999, 10, 0)

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

func TestCountTasksByJobNumber(t *testing.T) {
	Convey("Given a MongoDB connection with tasks for multiple jobs", t, func() {
		ctx := context.Background()
		m, conn := setupTaskStoreTest(t, ctx)
		collection := config.TasksCollectionName

		now := time.Now().UTC().Truncate(time.Millisecond)
		jobID := 888
		otherJobID := 777

		task1 := &domain.Task{
			ID:          "task-1",
			JobNumber:   jobID,
			State:       domain.StateSubmitted,
			LastUpdated: now,
		}
		task2 := &domain.Task{
			ID:          "task-2",
			JobNumber:   jobID,
			State:       domain.StateApproved,
			LastUpdated: now,
		}
		task3 := &domain.Task{
			ID:          "task-3",
			JobNumber:   jobID,
			State:       domain.StateCompleted,
			LastUpdated: now,
		}
		taskOtherJob := &domain.Task{
			ID:          "task-other-count",
			JobNumber:   otherJobID,
			State:       domain.StateSubmitted,
			LastUpdated: now,
		}

		testData := TaskList{task1, task2, task3, taskOtherJob}

		if err := setUpTestDataTasks(ctx, conn, collection, testData); err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}

		Convey("When counting tasks for different jobs", func() {
			count1, err1 := m.CountTasksByJobNumber(ctx, jobID)
			count2, err2 := m.CountTasksByJobNumber(ctx, otherJobID)
			count3, err3 := m.CountTasksByJobNumber(ctx, 99999)

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
			count1, err1 := m.CountTasksByJobNumber(ctx, jobID)
			count2, err2 := m.CountTasksByJobNumber(ctx, jobID)

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

func TestGetTask(t *testing.T) {
	Convey("Given a MongoDB connection with stored tasks", t, func() {
		ctx := context.Background()
		m, conn := setupTaskStoreTest(t, ctx)
		collection := config.TasksCollectionName

		task := domain.NewTask(123)
		task.State = domain.StateSubmitted
		conn.Collection(collection).InsertOne(ctx, task)

		Convey("When retrieving an existing task by ID", func() {
			retrieved, err := m.GetTask(ctx, task.ID)

			Convey("Then the task should be retrieved successfully", func() {
				So(err, ShouldBeNil)
				So(retrieved, ShouldNotBeNil)
				So(retrieved.ID, ShouldEqual, task.ID)
				So(retrieved.JobNumber, ShouldEqual, 123)
				So(retrieved.State, ShouldEqual, domain.StateSubmitted)
			})
		})

		Convey("When retrieving a non-existent task", func() {
			retrieved, err := m.GetTask(ctx, "non-existent-id")

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(retrieved, ShouldBeNil)
			})
		})

		Reset(func() {
			conn.DropDatabase(ctx)
		})
	})
}

func TestUpdateTask(t *testing.T) {
	Convey("Given a MongoDB connection with an existing task", t, func() {
		ctx := context.Background()
		m, conn := setupTaskStoreTest(t, ctx)
		collection := config.TasksCollectionName

		task := domain.NewTask(123)
		task.State = domain.StateSubmitted
		conn.Collection(collection).InsertOne(ctx, task)

		Convey("When updating the task", func() {
			task.State = domain.StateCompleted
			err := m.UpdateTask(ctx, &task)

			Convey("Then the task should be updated without error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the changes should be persisted in the database", func() {
				var retrieved domain.Task
				queryMongo(conn, collection, bson.M{"_id": task.ID}, &retrieved)
				So(retrieved.State, ShouldEqual, domain.StateCompleted)
			})
		})

		Convey("When updating a non-existent task", func() {
			nonExistent := domain.NewTask(999)
			err := m.UpdateTask(ctx, &nonExistent)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Reset(func() {
			conn.DropDatabase(ctx)
		})
	})
}

func TestUpdateTaskState(t *testing.T) {
	Convey("Given a MongoDB connection with an existing task", t, func() {
		ctx := context.Background()
		m, conn := setupTaskStoreTest(t, ctx)
		collection := config.TasksCollectionName

		task := domain.NewTask(123)
		task.State = domain.StateSubmitted
		originalTime := task.LastUpdated
		conn.Collection(collection).InsertOne(ctx, task)

		Convey("When updating the task state", func() {
			newTime := time.Now().Add(1 * time.Hour)
			err := m.UpdateTaskState(ctx, task.ID, domain.StateInReview, newTime)

			Convey("Then the task state should be updated without error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the state and timestamp should be updated in the database", func() {
				var retrieved domain.Task
				queryMongo(conn, collection, bson.M{"_id": task.ID}, &retrieved)
				So(retrieved.State, ShouldEqual, domain.StateInReview)
				So(retrieved.LastUpdated, ShouldHappenAfter, originalTime)
			})
		})

		Convey("When updating state for a non-existent task", func() {
			err := m.UpdateTaskState(ctx, "non-existent-id", domain.StateCompleted, time.Now())

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Reset(func() {
			conn.DropDatabase(ctx)
		})
	})
}

func TestClaimTask(t *testing.T) {
	Convey("Given a MongoDB connection with tasks in different states", t, func() {
		ctx := context.Background()
		m, conn := setupTaskStoreTest(t, ctx)
		collection := config.TasksCollectionName

		// Create tasks with different states
		task1 := domain.NewTask(123)
		task1.State = domain.StateSubmitted
		task1.Type = domain.TaskTypeDatasetSeries

		task2 := domain.NewTask(123)
		task2.State = domain.StateSubmitted
		task2.Type = domain.TaskTypeDatasetEdition

		task3 := domain.NewTask(123)
		task3.State = domain.StateInReview
		task3.Type = domain.TaskTypeDatasetVersion

		testData := TaskList{&task1, &task2, &task3}
		setUpTestDataTasks(ctx, conn, collection, testData)

		Convey("When claiming a submitted task", func() {
			claimed, err := m.ClaimTask(ctx, domain.StateSubmitted, domain.StateInReview)

			Convey("Then a submitted task should be claimed and updated", func() {
				So(err, ShouldBeNil)
				So(claimed, ShouldNotBeNil)
				So(claimed.State, ShouldEqual, domain.StateInReview)
			})

			Convey("And the task state should be updated in the database", func() {
				var retrieved domain.Task
				queryMongo(conn, collection, bson.M{"_id": claimed.ID}, &retrieved)
				So(retrieved.State, ShouldEqual, domain.StateInReview)
			})

			Convey("And claiming again should return another task", func() {
				claimed2, err2 := m.ClaimTask(ctx, domain.StateSubmitted, domain.StateInReview)
				So(err2, ShouldBeNil)
				So(claimed2, ShouldNotBeNil)
				So(claimed2.ID, ShouldNotEqual, claimed.ID)
			})
		})

		Convey("When no tasks match the pending state", func() {
			// Claim all submitted tasks first
			m.ClaimTask(ctx, domain.StateSubmitted, domain.StateInReview)
			m.ClaimTask(ctx, domain.StateSubmitted, domain.StateInReview)

			claimed, err := m.ClaimTask(ctx, domain.StateSubmitted, domain.StateInReview)

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
