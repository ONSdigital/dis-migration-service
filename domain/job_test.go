package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewJob(t *testing.T) {
	Convey("Given a valid job config and host", t, func() {
		jobConfig := JobConfig{
			Type:     JobTypeStaticDataset,
			SourceID: "/source-id",
			TargetID: "target-id",
		}
		label := "Test Dataset Title"

		Convey("When a job is created", func() {
			job := NewJob(&jobConfig, label)

			Convey("Then a valid job should be returned", func() {
				So(job.Config, ShouldResemble, &jobConfig)
				So(job.State, ShouldEqual, JobStateSubmitted)
				So(job.Label, ShouldEqual, label)
				So(uuid.Validate(job.ID), ShouldBeNil)
				So(job.Links.Self.HRef, ShouldEqual, fmt.Sprintf("/v1/migration-jobs/%s", job.ID))
				So(job.Links.Tasks.HRef, ShouldEqual, fmt.Sprintf("/v1/migration-jobs/%s/tasks", job.ID))
				So(job.Links.Events.HRef, ShouldEqual, fmt.Sprintf("/v1/migration-jobs/%s/events", job.ID))
				So(job.LastUpdated, ShouldHappenOnOrBetween, time.Now().Add(-5*time.Second), time.Now())
			})
		})
	})
}

func TestNewJobLinks(t *testing.T) {
	Convey("Given a valid id and host", t, func() {
		id := uuid.New().String()

		Convey("When a job links is created", func() {
			jobLinks := NewJobLinks(id)

			Convey("Then a valid jobLinks should be returned", func() {
				So(jobLinks.Self.HRef, ShouldEqual, fmt.Sprintf("/v1/migration-jobs/%s", id))
			})
		})
	})
}
