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

func TestCreateEvent(t *testing.T) {
	Convey("Given a mongo store with a mocked collection", t, func() {
		ctx := context.Background()
		mockCollection := &mock.MongoCollectionMock{
			InsertOneFunc: func(ctx context.Context, document interface{}) (*drivermongo.InsertOneResult, error) {
				return &drivermongo.InsertOneResult{InsertedID: "test-event-id"}, nil
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

		Convey("When creating an event successfully", func() {
			event := &domain.Event{
				ID:        "event-1",
				JobNumber: 1,
				Action:    "test_action",
				CreatedAt: time.Now().Format(time.RFC3339),
			}

			err := mongoStore.CreateEvent(ctx, event)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And InsertOne should be called once", func() {
				So(len(mockCollection.InsertOneCalls()), ShouldEqual, 1)
				So(mockCollection.InsertOneCalls()[0].Document, ShouldEqual, event)
			})
		})

		Convey("When InsertOne fails", func() {
			mockCollection.InsertOneFunc = func(ctx context.Context, document interface{}) (*drivermongo.InsertOneResult, error) {
				return nil, errors.New("insert error")
			}

			event := &domain.Event{ID: "event-1"}
			err := mongoStore.CreateEvent(ctx, event)

			Convey("Then ErrInternalServerError should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrInternalServerError)
			})
		})
	})
}

func TestGetJobEvents(t *testing.T) {
	Convey("Given a mongo store with a mocked collection", t, func() {
		ctx := context.Background()

		mockCollection := &mock.MongoCollectionMock{
			FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...interface{}) (int, error) {
				if events, ok := results.(*[]*domain.Event); ok {
					*events = []*domain.Event{
						{ID: "event-1", JobNumber: 123, Action: "event_1"},
						{ID: "event-2", JobNumber: 123, Action: "event_2"},
					}
				}
				return 15, nil
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

		Convey("When getting events for a job", func() {
			events, totalCount, err := mongoStore.GetJobEvents(ctx, 123, 10, 0)

			Convey("Then events should be returned", func() {
				So(err, ShouldBeNil)
				So(len(events), ShouldEqual, 2)
				So(totalCount, ShouldEqual, 15)
				So(events[0].JobNumber, ShouldEqual, 123)
			})

			Convey("And Find should be called once", func() {
				So(len(mockCollection.FindCalls()), ShouldEqual, 1)
			})
		})

		Convey("When Find fails", func() {
			mockCollection.FindFunc = func(ctx context.Context, filter interface{}, results interface{}, opts ...interface{}) (int, error) {
				return 0, errors.New("database error")
			}

			events, totalCount, err := mongoStore.GetJobEvents(ctx, 123, 10, 0)

			Convey("Then ErrInternalServerError should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrInternalServerError)
				So(events, ShouldBeNil)
				So(totalCount, ShouldEqual, 0)
			})
		})
	})
}

func TestCountEventsByJobNumber(t *testing.T) {
	Convey("Given a mongo store with a mocked collection", t, func() {
		ctx := context.Background()

		mockCollection := &mock.MongoCollectionMock{
			FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...interface{}) (int, error) {
				return 25, nil
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

		Convey("When counting events by job number", func() {
			count, err := mongoStore.CountEventsByJobNumber(ctx, 123)

			Convey("Then the count should be returned", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 25)
			})

			Convey("And Find should be called once", func() {
				So(len(mockCollection.FindCalls()), ShouldEqual, 1)
			})
		})

		Convey("When Find fails", func() {
			mockCollection.FindFunc = func(ctx context.Context, filter interface{}, results interface{}, opts ...interface{}) (int, error) {
				return 0, errors.New("database error")
			}

			count, err := mongoStore.CountEventsByJobNumber(ctx, 123)

			Convey("Then ErrInternalServerError should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrInternalServerError)
				So(count, ShouldEqual, 0)
			})
		})
	})
}
