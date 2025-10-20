package api

import (
	"testing"

	apiErrors "github.com/ONSdigital/dis-migration-service/api/errors"
	"github.com/ONSdigital/dis-migration-service/domain"
	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateJobConfig(t *testing.T) {
	Convey("Given a valid job config", t, func() {
		jobConfig := domain.JobConfig{
			SourceID: "source-id",
			TargetID: "target-id",
			Type:     "dataset",
		}

		Convey("When the config is validated", func() {
			errs := validateJobConfig(&jobConfig)

			Convey("Then no errors should be returend", func() {
				So(errs, ShouldBeNil)
			})
		})
	})

	Convey("Given a job config with a missing parameter", t, func() {
		jobConfig := domain.JobConfig{
			TargetID: "target-id",
			Type:     "dataset",
		}

		Convey("When the config is validated", func() {
			errs := validateJobConfig(&jobConfig)

			Convey("Then an error should be returend", func() {
				So(errs, ShouldHaveLength, 1)
				So(errs, ShouldContain, apiErrors.ErrSourceIDNotProvided)
			})
		})
	})

	Convey("Given a job config with a multiple missing parameters", t, func() {
		jobConfig := domain.JobConfig{}

		Convey("When the config is validated", func() {
			errs := validateJobConfig(&jobConfig)

			Convey("Then an error should be returend", func() {
				So(errs, ShouldHaveLength, 3)
				So(errs, ShouldContain, apiErrors.ErrSourceIDNotProvided)
				So(errs, ShouldContain, apiErrors.ErrTargetIDNotProvided)
				So(errs, ShouldContain, apiErrors.ErrJobTypeNotProvided)
			})
		})
	})
}
