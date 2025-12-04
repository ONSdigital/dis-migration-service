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

func TestNewTask(t *testing.T) {
	Convey("Given a job ID", t, func() {
		jobID := "job-456"

		Convey("When NewTask is called", func() {
			task := domain.NewTask(jobID)

			Convey("Then a Task should be returned with the correct fields set", func() {
				So(task.ID, ShouldNotBeEmpty)
				So(task.JobID, ShouldEqual, jobID)
				So(task.State, ShouldEqual, domain.TaskStateSubmitted)
				So(task.Links.Self.HRef, ShouldEqual,
					"/v1/migration-jobs/job-456/tasks/"+task.ID)
				So(task.Links.Job.HRef, ShouldEqual,
					"/v1/migration-jobs/job-456")
			})
		})
	})
}
