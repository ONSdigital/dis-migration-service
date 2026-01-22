package mongo_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dis-migration-service/mongo"
	"github.com/ONSdigital/dis-migration-service/mongo/mock"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	. "github.com/smartystreets/goconvey/convey"
	drivermongo "go.mongodb.org/mongo-driver/mongo"
)

func TestGetNextJobNumberCounter(t *testing.T) {
	Convey("Given a mongo store with a mocked collection", t, func() {
		ctx := context.Background()

		Convey("When the counter exists and is incremented successfully", func() {
			expectedCounter := &domain.Counter{
				CounterName:  "job_number_counter",
				CounterValue: 5,
			}

			mockCollection := &mock.MongoCollectionMock{
				FindOneAndUpdateFunc: func(ctx context.Context, filter interface{}, update interface{}, result interface{}, opts ...interface{}) error {
					if counter, ok := result.(*domain.Counter); ok {
						*counter = *expectedCounter
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

			counter, err := mongoStore.GetNextJobNumberCounter(ctx)

			Convey("Then the incremented counter should be returned", func() {
				So(err, ShouldBeNil)
				So(counter, ShouldNotBeNil)
				So(counter.CounterName, ShouldEqual, "job_number_counter")
				So(counter.CounterValue, ShouldEqual, 5)
			})

			Convey("And FindOneAndUpdate should be called once", func() {
				So(len(mockCollection.FindOneAndUpdateCalls()), ShouldEqual, 1)
			})
		})

		Convey("When the counter does not exist and needs to be created", func() {
			callCount := 0
			expectedCounter := &domain.Counter{
				CounterName:  "job_number_counter",
				CounterValue: 1,
			}

			mockCollection := &mock.MongoCollectionMock{
				FindOneAndUpdateFunc: func(ctx context.Context, filter interface{}, update interface{}, result interface{}, opts ...interface{}) error {
					callCount++
					if callCount == 1 {
						return mongodriver.ErrNoDocumentFound
					}
					if counter, ok := result.(*domain.Counter); ok {
						*counter = *expectedCounter
					}
					return nil
				},
				InsertOneFunc: func(ctx context.Context, document interface{}) (*drivermongo.InsertOneResult, error) {
					return &drivermongo.InsertOneResult{InsertedID: "counter-id"}, nil
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

			counter, err := mongoStore.GetNextJobNumberCounter(ctx)

			Convey("Then the counter should be created and incremented", func() {
				So(err, ShouldBeNil)
				So(counter, ShouldNotBeNil)
				So(counter.CounterValue, ShouldEqual, 1)
			})

			Convey("And InsertOne should be called once", func() {
				So(len(mockCollection.InsertOneCalls()), ShouldEqual, 1)
			})

			Convey("And FindOneAndUpdate should be called twice", func() {
				So(len(mockCollection.FindOneAndUpdateCalls()), ShouldEqual, 2)
			})
		})

		Convey("When creating the counter fails", func() {
			mockCollection := &mock.MongoCollectionMock{
				FindOneAndUpdateFunc: func(ctx context.Context, filter interface{}, update interface{}, result interface{}, opts ...interface{}) error {
					return mongodriver.ErrNoDocumentFound
				},
				InsertOneFunc: func(ctx context.Context, document interface{}) (*drivermongo.InsertOneResult, error) {
					return nil, errors.New("insert failed")
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

			counter, err := mongoStore.GetNextJobNumberCounter(ctx)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(counter, ShouldBeNil)
			})
		})

		Convey("When FindOneAndUpdate fails with a non-NotFound error", func() {
			mockCollection := &mock.MongoCollectionMock{
				FindOneAndUpdateFunc: func(ctx context.Context, filter interface{}, update interface{}, result interface{}, opts ...interface{}) error {
					return errors.New("database connection error")
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

			counter, err := mongoStore.GetNextJobNumberCounter(ctx)

			Convey("Then ErrInternalServerError should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrInternalServerError)
				So(counter, ShouldBeNil)
			})
		})
	})
}
