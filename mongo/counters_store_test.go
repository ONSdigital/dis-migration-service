package mongo_test

import (
	"context"
	"testing"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/mongo"
	mongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
)

func setupCounterStoreTest(t *testing.T, ctx context.Context) (*mongo.Mongo, *mongoDriver.MongoConnection) {
	conn, err := setupSharedMongo(t)
	if err != nil {
		t.Fatalf("failed to setup shared mongo: %v", err)
	}

	m := &mongo.Mongo{
		MongoConfig: config.MongoConfig{
			MongoDriverConfig: mongoDriver.MongoDriverConfig{
				Database: database,
				Collections: map[string]string{
					config.JobsCollectionTitle:     config.JobsCollectionName,
					config.EventsCollectionTitle:   config.EventsCollectionName,
					config.TasksCollectionTitle:    config.TasksCollectionName,
					config.CountersCollectionTitle: config.CountersCollectionName,
				},
			},
		},
		Connection: conn,
	}

	return m, conn
}

func TestGetNextJobNumberCounter(t *testing.T) {
	Convey("Given a MongoDB connection and counters collection", t, func() {
		ctx := context.Background()
		m, conn := setupCounterStoreTest(t, ctx)
		collection := config.CountersCollectionName

		Convey("When the counter does not exist", func() {
			// Ensure counter doesn't exist
			conn.Collection(collection).DeleteMany(ctx, bson.M{})

			counter, err := m.GetNextJobNumberCounter(ctx)

			Convey("Then a new counter should be created and returned with value 1", func() {
				So(err, ShouldBeNil)
				So(counter, ShouldNotBeNil)
				So(counter.CounterName, ShouldEqual, "job_number_counter")
				So(counter.CounterValue, ShouldEqual, 1)
			})

			Convey("And the counter should exist in the database", func() {
				var retrieved domain.Counter
				err := queryMongo(conn, collection, bson.M{"counter_name": "job_number_counter"}, &retrieved)
				So(err, ShouldBeNil)
				So(retrieved.CounterValue, ShouldEqual, 1)
			})

			Reset(func() {
				conn.DropDatabase(ctx)
			})
		})

		Convey("When the counter already exists", func() {
			// Setup: create initial counter
			initialCounter := domain.Counter{
				CounterName:  "job_number_counter",
				CounterValue: 5,
			}
			conn.Collection(collection).InsertOne(ctx, initialCounter)

			counter1, err1 := m.GetNextJobNumberCounter(ctx)

			Convey("Then it should be incremented to 6", func() {
				So(err1, ShouldBeNil)
				So(counter1, ShouldNotBeNil)
				So(counter1.CounterValue, ShouldEqual, 6)
			})

			Convey("And calling it again should increment to 7", func() {
				counter2, err2 := m.GetNextJobNumberCounter(ctx)
				So(err2, ShouldBeNil)
				So(counter2, ShouldNotBeNil)
				So(counter2.CounterValue, ShouldEqual, 7)
			})

			Convey("And multiple calls should increment sequentially", func() {
				counter2, _ := m.GetNextJobNumberCounter(ctx)
				counter3, _ := m.GetNextJobNumberCounter(ctx)
				counter4, _ := m.GetNextJobNumberCounter(ctx)

				So(counter2.CounterValue, ShouldEqual, 7)
				So(counter3.CounterValue, ShouldEqual, 8)
				So(counter4.CounterValue, ShouldEqual, 9)
			})

			Reset(func() {
				conn.DropDatabase(ctx)
			})
		})

		Convey("When multiple counters exist in the collection", func() {
			// Setup: create job counter and other counters
			conn.Collection(collection).InsertOne(ctx, domain.Counter{
				CounterName:  "job_number_counter",
				CounterValue: 10,
			})
			conn.Collection(collection).InsertOne(ctx, domain.Counter{
				CounterName:  "other_counter",
				CounterValue: 100,
			})

			counter, err := m.GetNextJobNumberCounter(ctx)

			Convey("Then only the job_number_counter should be incremented", func() {
				So(err, ShouldBeNil)
				So(counter.CounterValue, ShouldEqual, 11)

				// Verify other counter is unchanged
				var otherCounter domain.Counter
				queryMongo(conn, collection, bson.M{"counter_name": "other_counter"}, &otherCounter)
				So(otherCounter.CounterValue, ShouldEqual, 100)
			})

			Reset(func() {
				conn.DropDatabase(ctx)
			})
		})
	})
}
