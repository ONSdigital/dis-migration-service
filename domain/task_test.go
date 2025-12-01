package domain_test

import (
	"testing"

	"github.com/ONSdigital/dis-migration-service/domain"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewTaskLinks(t *testing.T) {
	Convey("Given a task ID and job ID", t, func() {
		id := "task-123"
		jobID := "job-789"

		Convey("When NewTaskLinks is called", func() {
			links := domain.NewTaskLinks(id, jobID)

			Convey("Then the Self link should be correctly constructed", func() {
				So(links.Self, ShouldNotBeNil)
				So(links.Self.HRef, ShouldEqual,
					"/v1/migration-jobs/job-789/tasks/task-123")
			})

			Convey("Then the Job link should be correctly constructed", func() {
				So(links.Job, ShouldNotBeNil)
				So(links.Job.HRef, ShouldEqual,
					"/v1/migration-jobs/job-789")
			})
		})
	})
}
