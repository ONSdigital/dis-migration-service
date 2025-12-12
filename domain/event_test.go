package domain_test

import (
	"testing"

	"github.com/ONSdigital/dis-migration-service/domain"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewEventLinks(t *testing.T) {
	Convey("Given an event ID and job ID", t, func() {
		id := "event-456"
		jobID := "job-123"

		Convey("When NewEventLinks is called", func() {
			links := domain.NewEventLinks(id, jobID)

			Convey("Then the Self link should be correctly constructed", func() {
				So(links.Self, ShouldNotBeNil)
				So(links.Self.HRef, ShouldEqual,
					"/v1/migration-jobs/job-123/events/event-456")
			})

			Convey("Then the Job link should be correctly constructed", func() {
				So(links.Job, ShouldNotBeNil)
				So(links.Job.HRef, ShouldEqual,
					"/v1/migration-jobs/job-123")
			})
		})
	})
}
