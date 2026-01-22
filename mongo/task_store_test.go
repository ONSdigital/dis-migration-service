package mongo_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dis-migration-service/mongo"
	"github.com/ONSdigital/dis-migration-service/mongo/mock"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	. "github.com/smartystreets/goconvey/convey"
	drivermongo "go.mongodb.org/mongo-driver/mongo"
)

func TestCreateTask(t *testing.T) {
	Convey("Given a mongo store with a mocked collection", t, func() {
		ctx := context.Background()
		mockCollection := &mock.MongoCollectionMock{
			InsertOneFunc: func(ctx context.Context, document interface{}) (*drivermongo.InsertOneResult, error) {
				return &drivermongo.InsertOneResult{InsertedID: "test-task-id"}, nil
			},
		}

		mockConnection := &mock.MongoConnectionMock{
			CollectionFunc: func(name string) mongo.MongoCollection {
				return mockCollection
			},
		}

		mongoStore := &mongo.Mongo{
			MongoConfig: config.MongoConfig{
				MongoDriverConfig: mongodriver.MongoDriverConfig{
					ClusterEndpoint: "localhost:27017",
					Database:        "test-db",
				},
			},
		}
		mongoStore.SetConnection(mockConnection)

		Convey("When creating a task successfully", func() {
			task := &domain.Task{
				ID:          "task-1",
				JobNumber:   1,
				State:       domain.StateSubmitted,
				LastUpdated: time.Now(),
			}

			err := mongoStore.CreateTask(ctx, task)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And InsertOne should be called once", func() {
				So(len(mockCollection.InsertOneCalls()), ShouldEqual, 1)
				So(mockCollection.InsertOneCalls()[0].Document, ShouldEqual, task)
			})
		})

		Convey("When InsertOne fails", func() {
			mockCollection.InsertOneFunc = func(ctx context.Context, document interface{}) (*drivermongo.InsertOneResult, error) {
				return nil, errors.New("insert error")
			}

			task := &domain.Task{ID: "task-1"}
			err := mongoStore.CreateTask(ctx, task)

			Convey("Then ErrInternalServerError should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrInternalServerError)
			})
		})
	})
}

func TestGetTask(t *testing.T) {
	Convey("Given a mongo store with a mocked collection", t, func() {
		ctx := context.Background()
		expectedTask := &domain.Task{
			ID:        "task-1",
			JobNumber: 123,
			State:     domain.StateSubmitted,
		}

		mockCollection := &mock.MongoCollectionMock{
			FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...interface{}) error {
				if task, ok := result.(*domain.Task); ok {
					*task = *expectedTask
				}
				return nil
			},
		}

		mockConnection := &mock.MongoConnectionMock{
			CollectionFunc: func(name string) mongo.MongoCollection {
				return mockCollection
			},
		}

		mongoStore := &mongo.Mongo{
			MongoConfig: config.MongoConfig{
				MongoDriverConfig: mongodriver.MongoDriverConfig{
					ClusterEndpoint: "localhost:27017",
					Database:        "test-db",
				},
			},
		}
		mongoStore.SetConnection(mockConnection)

		Convey("When getting a task that exists", func() {
			task, err := mongoStore.GetTask(ctx, "task-1")

			Convey("Then the task should be returned", func() {
				So(err, ShouldBeNil)
				So(task, ShouldNotBeNil)
				So(task.ID, ShouldEqual, "task-1")
				So(task.JobNumber, ShouldEqual, 123)
			})

			Convey("And FindOne should be called once", func() {
				So(len(mockCollection.FindOneCalls()), ShouldEqual, 1)
			})
		})

		Convey("When the task is not found", func() {
			mockCollection.FindOneFunc = func(ctx context.Context, filter interface{}, result interface{}, opts ...interface{}) error {
				return mongodriver.ErrNoDocumentFound
			}

			task, err := mongoStore.GetTask(ctx, "task-999")

			Convey("Then ErrTaskNotFound should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrTaskNotFound)
				So(task, ShouldBeNil)
			})
		})

		Convey("When FindOne returns a different error", func() {
			mockCollection.FindOneFunc = func(ctx context.Context, filter interface{}, result interface{}, opts ...interface{}) error {
				return errors.New("database error")
			}

			task, err := mongoStore.GetTask(ctx, "task-1")

			Convey("Then ErrInternalServerError should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrInternalServerError)
				So(task, ShouldBeNil)
			})
		})
	})
}

func TestUpdateTask(t *testing.T) {
	Convey("Given a mongo store with a mocked collection", t, func() {
		ctx := context.Background()

		mockCollection := &mock.MongoCollectionMock{
			UpdateOneFunc: func(ctx context.Context, filter interface{}, update interface{}, opts ...interface{}) (*drivermongo.UpdateResult, error) {
				return &drivermongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil
			},
		}

		mockConnection := &mock.MongoConnectionMock{
			CollectionFunc: func(name string) mongo.MongoCollection {
				return mockCollection
			},
		}

		mongoStore := &mongo.Mongo{
			MongoConfig: config.MongoConfig{
				MongoDriverConfig: mongodriver.MongoDriverConfig{
					ClusterEndpoint: "localhost:27017",
					Database:        "test-db",
				},
			},
		}
		mongoStore.SetConnection(mockConnection)

		Convey("When updating a task successfully", func() {
			task := &domain.Task{
				ID:    "task-1",
				State: domain.StateCompleted,
			}

			err := mongoStore.UpdateTask(ctx, task)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And UpdateOne should be called once", func() {
				So(len(mockCollection.UpdateOneCalls()), ShouldEqual, 1)
			})
		})

		Convey("When the task is not found", func() {
			mockCollection.UpdateOneFunc = func(ctx context.Context, filter interface{}, update interface{}, opts ...interface{}) (*drivermongo.UpdateResult, error) {
				return &drivermongo.UpdateResult{MatchedCount: 0}, nil
			}

			task := &domain.Task{ID: "task-999"}
			err := mongoStore.UpdateTask(ctx, task)

			Convey("Then ErrTaskNotFound should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrTaskNotFound)
			})
		})

		Convey("When UpdateOne fails", func() {
			mockCollection.UpdateOneFunc = func(ctx context.Context, filter interface{}, update interface{}, opts ...interface{}) (*drivermongo.UpdateResult, error) {
				return nil, errors.New("database error")
			}

			task := &domain.Task{ID: "task-1"}
			err := mongoStore.UpdateTask(ctx, task)

			Convey("Then ErrInternalServerError should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrInternalServerError)
			})
		})
	})
}

func TestUpdateTaskState(t *testing.T) {
	Convey("Given a mongo store with a mocked collection", t, func() {
		ctx := context.Background()
		now := time.Now()

		mockCollection := &mock.MongoCollectionMock{
			UpdateOneFunc: func(ctx context.Context, filter interface{}, update interface{}, opts ...interface{}) (*drivermongo.UpdateResult, error) {
				return &drivermongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil
			},
		}

		mockConnection := &mock.MongoConnectionMock{
			CollectionFunc: func(name string) mongo.MongoCollection {
				return mockCollection
			},
		}

		mongoStore := &mongo.Mongo{
			MongoConfig: config.MongoConfig{
				MongoDriverConfig: mongodriver.MongoDriverConfig{
					ClusterEndpoint: "localhost:27017",
					Database:        "test-db",
				},
			},
		}
		mongoStore.SetConnection(mockConnection)

		Convey("When updating task state successfully", func() {
			err := mongoStore.UpdateTaskState(ctx, "task-1", domain.StateCompleted, now)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And UpdateOne should be called once", func() {
				So(len(mockCollection.UpdateOneCalls()), ShouldEqual, 1)
			})
		})

		Convey("When the task is not found", func() {
			mockCollection.UpdateOneFunc = func(ctx context.Context, filter interface{}, update interface{}, opts ...interface{}) (*drivermongo.UpdateResult, error) {
				return &drivermongo.UpdateResult{MatchedCount: 0}, nil
			}

			err := mongoStore.UpdateTaskState(ctx, "task-999", domain.StateCompleted, now)

			Convey("Then ErrTaskNotFound should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrTaskNotFound)
			})
		})

		Convey("When UpdateOne fails", func() {
			mockCollection.UpdateOneFunc = func(ctx context.Context, filter interface{}, update interface{}, opts ...interface{}) (*drivermongo.UpdateResult, error) {
				return nil, errors.New("database error")
			}

			err := mongoStore.UpdateTaskState(ctx, "task-1", domain.StateCompleted, now)

			Convey("Then ErrInternalServerError should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrInternalServerError)
			})
		})
	})
}

func TestClaimTask(t *testing.T) {
	Convey("Given a mongo store with a mocked collection", t, func() {
		ctx := context.Background()
		expectedTask := &domain.Task{
			ID:        "task-1",
			JobNumber: 1,
			State:     domain.StateMigrating,
		}

		mockCollection := &mock.MongoCollectionMock{
			FindOneAndUpdateFunc: func(ctx context.Context, filter interface{}, update interface{}, result interface{}, opts ...interface{}) error {
				if task, ok := result.(*domain.Task); ok {
					*task = *expectedTask
				}
				return nil
			},
		}

		mockConnection := &mock.MongoConnectionMock{
			CollectionFunc: func(name string) mongo.MongoCollection {
				return mockCollection
			},
		}

		mongoStore := &mongo.Mongo{
			MongoConfig: config.MongoConfig{
				MongoDriverConfig: mongodriver.MongoDriverConfig{
					ClusterEndpoint: "localhost:27017",
					Database:        "test-db",
				},
			},
		}
		mongoStore.SetConnection(mockConnection)

		Convey("When claiming a submitted task successfully", func() {
			task, err := mongoStore.ClaimTask(ctx, domain.StateSubmitted, domain.StateMigrating)

			Convey("Then the claimed task should be returned", func() {
				So(err, ShouldBeNil)
				So(task, ShouldNotBeNil)
				So(task.State, ShouldEqual, domain.StateMigrating)
			})

			Convey("And FindOneAndUpdate should be called once", func() {
				So(len(mockCollection.FindOneAndUpdateCalls()), ShouldEqual, 1)
			})
		})

		Convey("When no submitted task is found", func() {
			mockCollection.FindOneAndUpdateFunc = func(ctx context.Context, filter interface{}, update interface{}, result interface{}, opts ...interface{}) error {
				return mongodriver.ErrNoDocumentFound
			}

			task, err := mongoStore.ClaimTask(ctx, domain.StateSubmitted, domain.StateMigrating)

			Convey("Then no error and no task should be returned", func() {
				So(err, ShouldBeNil)
				So(task, ShouldBeNil)
			})
		})

		Convey("When FindOneAndUpdate returns an error", func() {
			mockCollection.FindOneAndUpdateFunc = func(ctx context.Context, filter interface{}, update interface{}, result interface{}, opts ...interface{}) error {
				return errors.New("database error")
			}

			task, err := mongoStore.ClaimTask(ctx, domain.StateSubmitted, domain.StateMigrating)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(task, ShouldBeNil)
			})
		})
	})
}

func TestGetJobTasks(t *testing.T) {
	Convey("Given a mongo store with a mocked collection", t, func() {
		ctx := context.Background()

		mockCollection := &mock.MongoCollectionMock{
			FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...interface{}) (int, error) {
				if tasks, ok := results.(*[]*domain.Task); ok {
					*tasks = []*domain.Task{
						{ID: "task-1", JobNumber: 123},
						{ID: "task-2", JobNumber: 123},
					}
				}
				return 5, nil
			},
		}

		mockConnection := &mock.MongoConnectionMock{
			CollectionFunc: func(name string) mongo.MongoCollection {
				return mockCollection
			},
		}

		mongoStore := &mongo.Mongo{
			MongoConfig: config.MongoConfig{
				MongoDriverConfig: mongodriver.MongoDriverConfig{
					ClusterEndpoint: "localhost:27017",
					Database:        "test-db",
				},
			},
		}
		mongoStore.SetConnection(mockConnection)

		Convey("When getting tasks with state filter", func() {
			tasks, totalCount, err := mongoStore.GetJobTasks(ctx, []domain.State{domain.StateSubmitted}, 123, 10, 0)

			Convey("Then tasks should be returned", func() {
				So(err, ShouldBeNil)
				So(len(tasks), ShouldEqual, 2)
				So(totalCount, ShouldEqual, 5)
			})

			Convey("And Find should be called once", func() {
				So(len(mockCollection.FindCalls()), ShouldEqual, 1)
			})
		})

		Convey("When Find fails", func() {
			mockCollection.FindFunc = func(ctx context.Context, filter interface{}, results interface{}, opts ...interface{}) (int, error) {
				return 0, errors.New("database error")
			}

			tasks, totalCount, err := mongoStore.GetJobTasks(ctx, []domain.State{}, 123, 10, 0)

			Convey("Then ErrInternalServerError should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrInternalServerError)
				So(tasks, ShouldBeNil)
				So(totalCount, ShouldEqual, 0)
			})
		})
	})
}

func TestCountTasksByJobNumber(t *testing.T) {
	Convey("Given a mongo store with a mocked collection", t, func() {
		ctx := context.Background()

		mockCollection := &mock.MongoCollectionMock{
			FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...interface{}) (int, error) {
				return 42, nil
			},
		}

		mockConnection := &mock.MongoConnectionMock{
			CollectionFunc: func(name string) mongo.MongoCollection {
				return mockCollection
			},
		}

		mongoStore := &mongo.Mongo{
			MongoConfig: config.MongoConfig{
				MongoDriverConfig: mongodriver.MongoDriverConfig{
					ClusterEndpoint: "localhost:27017",
					Database:        "test-db",
				},
			},
		}
		mongoStore.SetConnection(mockConnection)

		Convey("When counting tasks by job number", func() {
			count, err := mongoStore.CountTasksByJobNumber(ctx, 123)

			Convey("Then the count should be returned", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 42)
			})

			Convey("And Find should be called once", func() {
				So(len(mockCollection.FindCalls()), ShouldEqual, 1)
			})
		})

		Convey("When Find fails", func() {
			mockCollection.FindFunc = func(ctx context.Context, filter interface{}, results interface{}, opts ...interface{}) (int, error) {
				return 0, errors.New("database error")
			}

			count, err := mongoStore.CountTasksByJobNumber(ctx, 123)

			Convey("Then ErrInternalServerError should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrInternalServerError)
				So(count, ShouldEqual, 0)
			})
		})
	})
}
