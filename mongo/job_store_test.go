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

func TestCreateJob(t *testing.T) {
	Convey("Given a mongo store with a mocked collection", t, func() {
		ctx := context.Background()
		mockCollection := &mock.MongoCollectionMock{
			InsertOneFunc: func(ctx context.Context, document interface{}) (*drivermongo.InsertOneResult, error) {
				return &drivermongo.InsertOneResult{InsertedID: "test-id"}, nil
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

		Convey("When creating a job successfully", func() {
			job := &domain.Job{
				ID:          "job-1",
				JobNumber:   1,
				State:       domain.StateSubmitted,
				LastUpdated: time.Now(),
			}

			err := mongoStore.CreateJob(ctx, job)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And InsertOne should be called once", func() {
				So(len(mockCollection.InsertOneCalls()), ShouldEqual, 1)
				So(mockCollection.InsertOneCalls()[0].Document, ShouldEqual, job)
			})
		})

		Convey("When InsertOne fails", func() {
			mockCollection.InsertOneFunc = func(ctx context.Context, document interface{}) (*drivermongo.InsertOneResult, error) {
				return nil, errors.New("insert error")
			}

			job := &domain.Job{ID: "job-1"}
			err := mongoStore.CreateJob(ctx, job)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestGetJob(t *testing.T) {
	Convey("Given a mongo store with a mocked collection", t, func() {
		ctx := context.Background()
		expectedJob := &domain.Job{
			ID:        "job-1",
			JobNumber: 123,
			State:     domain.StateSubmitted,
		}

		mockCollection := &mock.MongoCollectionMock{
			FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...interface{}) error {
				if job, ok := result.(*domain.Job); ok {
					*job = *expectedJob
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
				},
			},
		}
		mongoStore.SetConnection(mockConnection)

		Convey("When getting a job that exists", func() {
			job, err := mongoStore.GetJob(ctx, 123)

			Convey("Then the job should be returned", func() {
				So(err, ShouldBeNil)
				So(job, ShouldNotBeNil)
				So(job.ID, ShouldEqual, "job-1")
				So(job.JobNumber, ShouldEqual, 123)
			})

			Convey("And FindOne should be called once", func() {
				So(len(mockCollection.FindOneCalls()), ShouldEqual, 1)
			})
		})

		Convey("When the job is not found", func() {
			mockCollection.FindOneFunc = func(ctx context.Context, filter interface{}, result interface{}, opts ...interface{}) error {
				return mongodriver.ErrNoDocumentFound
			}

			job, err := mongoStore.GetJob(ctx, 999)

			Convey("Then ErrJobNotFound should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrJobNotFound)
				So(job, ShouldBeNil)
			})
		})

		Convey("When FindOne returns a different error", func() {
			mockCollection.FindOneFunc = func(ctx context.Context, filter interface{}, result interface{}, opts ...interface{}) error {
				return errors.New("database error")
			}

			job, err := mongoStore.GetJob(ctx, 123)

			Convey("Then ErrInternalServerError should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrInternalServerError)
				So(job, ShouldBeNil)
			})
		})
	})
}

func TestUpdateJob(t *testing.T) {
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
				},
			},
		}
		mongoStore.SetConnection(mockConnection)

		Convey("When updating a job successfully", func() {
			job := &domain.Job{
				ID:    "job-1",
				State: domain.StateCompleted,
			}

			err := mongoStore.UpdateJob(ctx, job)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And UpdateOne should be called once", func() {
				So(len(mockCollection.UpdateOneCalls()), ShouldEqual, 1)
			})
		})

		Convey("When the job is not found", func() {
			mockCollection.UpdateOneFunc = func(ctx context.Context, filter interface{}, update interface{}, opts ...interface{}) (*drivermongo.UpdateResult, error) {
				return &drivermongo.UpdateResult{MatchedCount: 0}, nil
			}

			job := &domain.Job{ID: "job-999"}
			err := mongoStore.UpdateJob(ctx, job)

			Convey("Then ErrJobNotFound should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrJobNotFound)
			})
		})

		Convey("When UpdateOne fails", func() {
			mockCollection.UpdateOneFunc = func(ctx context.Context, filter interface{}, update interface{}, opts ...interface{}) (*drivermongo.UpdateResult, error) {
				return nil, errors.New("database error")
			}

			job := &domain.Job{ID: "job-1"}
			err := mongoStore.UpdateJob(ctx, job)

			Convey("Then ErrInternalServerError should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrInternalServerError)
			})
		})
	})
}
