package domain_test

import (
	"testing"

	"github.com/ONSdigital/dis-migration-service/domain"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewEvent(t *testing.T) {
	Convey("Given domain event creation", t, func() {
		jobNumber := 123
		action := "approved"
		userID := "user-456"

		Convey("When NewEvent is called with valid parameters", func() {
			event := domain.NewEvent(jobNumber, action, userID)

			Convey("Then an event should be created with correct values", func() {
				So(event, ShouldNotBeNil)
				So(event.JobNumber, ShouldEqual, jobNumber)
				So(event.Action, ShouldEqual, action)
				So(event.RequestedBy.ID, ShouldEqual, userID)
				So(event.ID, ShouldNotBeEmpty)
				So(event.CreatedAt, ShouldNotBeEmpty)
				So(event.Links, ShouldNotBeNil)
			})
		})

		Convey("When NewEvent is called with empty user ID", func() {
			event := domain.NewEvent(jobNumber, action, "")

			Convey("Then the user ID should default to 'system'", func() {
				So(event.RequestedBy.ID, ShouldEqual, "system")
			})
		})
	})
}

func TestNewEventLinks(t *testing.T) {
	Convey("Given an event ID and job ID", t, func() {
		id := "event-456"
		jobNumber := "123"

		Convey("When NewEventLinks is called", func() {
			links := domain.NewEventLinks(id, jobNumber)

			Convey("Then the Self link should be correctly constructed", func() {
				So(links.Self, ShouldNotBeNil)
				So(links.Self.HRef, ShouldEqual,
					"/v1/migration-jobs/123/events/event-456")
			})

			Convey("Then the Job link should be correctly constructed", func() {
				So(links.Job, ShouldNotBeNil)
				So(links.Job.HRef, ShouldEqual,
					"/v1/migration-jobs/123")
			})
		})
	})
}
