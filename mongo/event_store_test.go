package mongo_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/mongo"
	mongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
)

type EventList []*domain.Event

func (el EventList) AsInterfaceList() []interface{} {
	result := make([]interface{}, len(el))
	for i, event := range el {
		result[i] = event
	}
	return result
}

func setupEventStoreTest(t *testing.T, ctx context.Context) (*mongo.Mongo, *mongoDriver.MongoConnection) {
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

func TestCreateEvent(t *testing.T) {
	Convey("Given a MongoDB connection and events collection", t, func() {
		ctx := context.Background()
		m, conn := setupEventStoreTest(t, ctx)
		collection := config.EventsCollectionName

		Convey("When creating a new event with user information", func() {
			user := &domain.User{
				ID:    "user-123",
				Email: "user@example.com",
			}
			links := domain.NewEventLinks("event-1", "123")
			event := &domain.Event{
				ID:          "event-1",
				JobNumber:   123,
				Action:      "migration_started",
				CreatedAt:   time.Now().UTC().String(),
				RequestedBy: user,
				Links:       links,
			}

			err := m.CreateEvent(ctx, event)

			Convey("Then the event should be created without error", func() {
				So(err, ShouldBeNil)
				var retrieved domain.Event
				err := queryMongo(conn, collection, bson.M{"_id": event.ID}, &retrieved)
				So(err, ShouldBeNil)
				So(retrieved.ID, ShouldEqual, event.ID)
				So(retrieved.JobNumber, ShouldEqual, 123)
				So(retrieved.Action, ShouldEqual, "migration_started")
			})

			Convey("And the user information should be preserved", func() {
				var retrieved domain.Event
				queryMongo(conn, collection, bson.M{"_id": event.ID}, &retrieved)
				So(retrieved.RequestedBy, ShouldNotBeNil)
				So(retrieved.RequestedBy.ID, ShouldEqual, "user-123")
				So(retrieved.RequestedBy.Email, ShouldEqual, "user@example.com")
			})

			Convey("And the event links should be preserved", func() {
				var retrieved domain.Event
				queryMongo(conn, collection, bson.M{"_id": event.ID}, &retrieved)
				So(retrieved.Links.Self, ShouldNotBeNil)
				So(retrieved.Links.Self.HRef, ShouldEqual, "/v1/migration-jobs/123/events/event-1")
				So(retrieved.Links.Job, ShouldNotBeNil)
				So(retrieved.Links.Job.HRef, ShouldEqual, "/v1/migration-jobs/123")
			})

			Reset(func() {
				conn.DropDatabase(ctx)
			})
		})

		Convey("When creating an event without user information", func() {
			links := domain.NewEventLinks("event-2", "456")
			event := &domain.Event{
				ID:        "event-2",
				JobNumber: 456,
				Action:    "migration_paused",
				CreatedAt: time.Now().UTC().String(),
				Links:     links,
			}

			err := m.CreateEvent(ctx, event)

			Convey("Then the event should be created without error", func() {
				So(err, ShouldBeNil)
				var retrieved domain.Event
				queryMongo(conn, collection, bson.M{"_id": event.ID}, &retrieved)
				So(retrieved.ID, ShouldEqual, event.ID)
				So(retrieved.RequestedBy, ShouldBeNil)
			})

			Reset(func() {
				conn.DropDatabase(ctx)
			})
		})
	})
}

func TestGetJobEvents(t *testing.T) {
	Convey("Given a MongoDB connection with events for multiple jobs", t, func() {
		ctx := context.Background()
		m, conn := setupEventStoreTest(t, ctx)
		collection := config.EventsCollectionName

		now := time.Now().UTC()
		jobID := 123
		otherJobID := 456

		user1 := &domain.User{
			ID:    "user-1",
			Email: "user1@example.com",
		}
		user2 := &domain.User{
			ID:    "user-2",
			Email: "user2@example.com",
		}

		event1 := &domain.Event{
			ID:          "event-1",
			JobNumber:   jobID,
			Action:      "migration_started",
			CreatedAt:   now.Add(-3 * time.Hour).String(),
			RequestedBy: user1,
			Links:       domain.NewEventLinks("event-1", strconv.Itoa(jobID)),
		}
		event2 := &domain.Event{
			ID:          "event-2",
			JobNumber:   jobID,
			Action:      "migration_processing",
			CreatedAt:   now.Add(-2 * time.Hour).String(),
			RequestedBy: user2,
			Links:       domain.NewEventLinks("event-2", strconv.Itoa(jobID)),
		}
		event3 := &domain.Event{
			ID:          "event-3",
			JobNumber:   jobID,
			Action:      "migration_validating",
			CreatedAt:   now.Add(-1 * time.Hour).String(),
			RequestedBy: user1,
			Links:       domain.NewEventLinks("event-3", strconv.Itoa(jobID)),
		}
		event4 := &domain.Event{
			ID:          "event-4",
			JobNumber:   jobID,
			Action:      "migration_completed",
			CreatedAt:   now.String(),
			RequestedBy: user2,
			Links:       domain.NewEventLinks("event-4", strconv.Itoa(jobID)),
		}
		eventOtherJob := &domain.Event{
			ID:          "event-other",
			JobNumber:   otherJobID,
			Action:      "migration_started",
			CreatedAt:   now.String(),
			RequestedBy: user1,
			Links:       domain.NewEventLinks("event-other", strconv.Itoa(otherJobID)),
		}

		testData := EventList{event1, event2, event3, event4, eventOtherJob}

		if err := setUpTestDataEvents(ctx, conn, collection, testData); err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}

		Convey("When retrieving all events for a job without pagination", func() {
			retrieved, count, err := m.GetJobEvents(ctx, jobID, 10, 0)

			Convey("Then the operation should succeed with correct data and sorting", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 4)
				So(len(retrieved), ShouldEqual, 4)
				for _, event := range retrieved {
					So(event.JobNumber, ShouldEqual, jobID)
					So(event.ID, ShouldNotEqual, "event-other")
				}
				// Verify sorting by created_at descending (newest first)
				So(retrieved[0].ID, ShouldEqual, event4.ID) // Most recent
				So(retrieved[3].ID, ShouldEqual, event1.ID) // Oldest
			})

			Convey("And all user information should be preserved", func() {
				So(retrieved[0].RequestedBy, ShouldNotBeNil)
				So(retrieved[0].RequestedBy.ID, ShouldEqual, "user-2")
				So(retrieved[3].RequestedBy.ID, ShouldEqual, "user-1")
			})

			Convey("And all event links should be correct", func() {
				for _, event := range retrieved {
					So(event.Links.Self, ShouldNotBeNil)
					So(event.Links.Job, ShouldNotBeNil)
					So(event.Links.Self.HRef, ShouldContainSubstring, fmt.Sprintf("/v1/migration-jobs/%d", jobID))
					So(event.Links.Job.HRef, ShouldEqual, fmt.Sprintf("/v1/migration-jobs/%d", jobID))
				}
			})
		})

		Convey("When retrieving events with limit parameter", func() {
			retrieved, count, err := m.GetJobEvents(ctx, jobID, 2, 0)

			Convey("Then the operation should succeed with limited results", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 4)          // Total count
				So(len(retrieved), ShouldEqual, 2) // Limited results
			})
		})

		Convey("When retrieving events with offset parameter", func() {
			retrieved, count, err := m.GetJobEvents(ctx, jobID, 10, 2)

			Convey("Then the operation should succeed with offset applied", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 4)
				So(len(retrieved), ShouldEqual, 2) // 4 total - 2 offset
			})
		})

		Convey("When retrieving events with pagination", func() {
			page1, count1, err1 := m.GetJobEvents(ctx, jobID, 2, 0)
			page2, count2, err2 := m.GetJobEvents(ctx, jobID, 2, 2)

			Convey("Then pagination should work correctly with no overlaps", func() {
				So(err1, ShouldBeNil)
				So(err2, ShouldBeNil)
				So(count1, ShouldEqual, 4)
				So(count2, ShouldEqual, 4)
				So(len(page1), ShouldEqual, 2)
				So(len(page2), ShouldEqual, 2)
				So(page1[0].ID, ShouldNotEqual, page2[0].ID)
				So(page1[1].ID, ShouldNotEqual, page2[1].ID)
			})
		})

		Convey("When retrieving events for a non-existent job", func() {
			retrieved, count, err := m.GetJobEvents(ctx, 99999, 10, 0)

			Convey("Then an empty list should be returned", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 0)
				So(len(retrieved), ShouldEqual, 0)
			})
		})

		Convey("When retrieving events with different actions", func() {
			retrieved, count, err := m.GetJobEvents(ctx, jobID, 10, 0)

			Convey("Then all event actions should be preserved", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 4)

				actions := make(map[string]bool)
				for _, event := range retrieved {
					actions[event.Action] = true
				}
				So(actions["migration_started"], ShouldBeTrue)
				So(actions["migration_processing"], ShouldBeTrue)
				So(actions["migration_validating"], ShouldBeTrue)
				So(actions["migration_completed"], ShouldBeTrue)
			})
		})

		Reset(func() {
			conn.DropDatabase(ctx)
		})
	})
}

func TestCountEventsByJobNumber(t *testing.T) {
	Convey("Given a MongoDB connection with events for multiple jobs", t, func() {
		ctx := context.Background()
		m, conn := setupEventStoreTest(t, ctx)
		collection := config.EventsCollectionName

		now := time.Now().UTC().String()
		jobID := 789
		otherJobID := 999

		user := &domain.User{
			ID:    "user-admin",
			Email: "admin@example.com",
		}

		event1 := &domain.Event{
			ID:          "event-1",
			JobNumber:   jobID,
			Action:      "migration_started",
			CreatedAt:   now,
			RequestedBy: user,
			Links:       domain.NewEventLinks("event-1", strconv.Itoa(jobID)),
		}
		event2 := &domain.Event{
			ID:          "event-2",
			JobNumber:   jobID,
			Action:      "migration_processing",
			CreatedAt:   now,
			RequestedBy: user,
			Links:       domain.NewEventLinks("event-2", strconv.Itoa(jobID)),
		}
		event3 := &domain.Event{
			ID:          "event-3",
			JobNumber:   jobID,
			Action:      "migration_completed",
			CreatedAt:   now,
			RequestedBy: user,
			Links:       domain.NewEventLinks("event-3", strconv.Itoa(jobID)),
		}
		event4 := &domain.Event{
			ID:          "event-4",
			JobNumber:   jobID,
			Action:      "migration_completed",
			CreatedAt:   now,
			RequestedBy: user,
			Links:       domain.NewEventLinks("event-4", strconv.Itoa(jobID)),
		}
		eventOtherJob := &domain.Event{
			ID:          "event-other",
			JobNumber:   otherJobID,
			Action:      "migration_started",
			CreatedAt:   now,
			RequestedBy: user,
			Links:       domain.NewEventLinks("event-other", strconv.Itoa(otherJobID)),
		}

		testData := EventList{event1, event2, event3, event4, eventOtherJob}

		if err := setUpTestDataEvents(ctx, conn, collection, testData); err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}

		Convey("When counting events for a job with multiple events", func() {
			count, err := m.CountEventsByJobNumber(ctx, jobID)

			Convey("Then the operation should succeed with correct count", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 4)
			})
		})

		Convey("When counting events for a job with single event", func() {
			count, err := m.CountEventsByJobNumber(ctx, otherJobID)

			Convey("Then the operation should succeed with correct count", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})
		})

		Convey("When counting events for a job with no events", func() {
			count, err := m.CountEventsByJobNumber(ctx, 99999)

			Convey("Then the operation should succeed with zero count", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 0)
			})
		})

		Convey("When counting events for multiple different jobs", func() {
			count1, err1 := m.CountEventsByJobNumber(ctx, jobID)
			count2, err2 := m.CountEventsByJobNumber(ctx, otherJobID)
			count3, err3 := m.CountEventsByJobNumber(ctx, 99999)

			Convey("Then counts should be correct for each job", func() {
				So(err1, ShouldBeNil)
				So(err2, ShouldBeNil)
				So(err3, ShouldBeNil)
				So(count1, ShouldEqual, 4)
				So(count2, ShouldEqual, 1)
				So(count3, ShouldEqual, 0)
				So(count1, ShouldNotEqual, count2)
			})
		})

		Convey("When counting events for the same job multiple times", func() {
			count1, err1 := m.CountEventsByJobNumber(ctx, jobID)
			count2, err2 := m.CountEventsByJobNumber(ctx, jobID)

			Convey("Then counts should be consistent", func() {
				So(err1, ShouldBeNil)
				So(err2, ShouldBeNil)
				So(count1, ShouldEqual, count2)
				So(count1, ShouldEqual, 4)
			})
		})

		Convey("When counting events does not affect other jobs", func() {
			// Count one job
			count1Before, _ := m.CountEventsByJobNumber(ctx, jobID)

			// Count other job
			countOther, _ := m.CountEventsByJobNumber(ctx, otherJobID)

			// Count first job again
			count1After, _ := m.CountEventsByJobNumber(ctx, jobID)

			Convey("Then counts should remain stable and isolated", func() {
				So(count1Before, ShouldEqual, count1After)
				So(count1Before, ShouldEqual, 4)
				So(countOther, ShouldEqual, 1)
			})
		})

		Reset(func() {
			conn.DropDatabase(ctx)
		})
	})
}

func setUpTestDataEvents(ctx context.Context, mongoConnection *mongoDriver.MongoConnection, collection string, events EventList) error {
	if err := mongoConnection.DropDatabase(ctx); err != nil {
		return err
	}

	if _, err := mongoConnection.Collection(collection).InsertMany(
		ctx,
		events.AsInterfaceList(),
	); err != nil {
		return err
	}

	return nil
}
