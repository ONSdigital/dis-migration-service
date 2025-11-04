package domain

import (
	"testing"

	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateJobType(t *testing.T) {
	Convey("Given a valid job type", t, func() {
		jobType := JobTypeStaticDataset

		Convey("When the type is validated", func() {
			err := jobType.Validate()

			Convey("Then no errors should be returend", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a invalid job type", t, func() {
		const faketype JobType = "faketype"

		Convey("When the type is validated", func() {
			err := faketype.Validate()

			Convey("Then an error should be returend", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, appErrors.ErrJobTypeInvalid)
			})
		})
	})

	Convey("Given a blank job type", t, func() {
		const faketype JobType = ""

		Convey("When the type is validated", func() {
			err := faketype.Validate()

			Convey("Then an error should be returend", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, appErrors.ErrJobTypeNotProvided)
			})
		})
	})
}
