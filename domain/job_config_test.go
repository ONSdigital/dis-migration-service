package domain

import (
	"testing"

	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateJobConfig(t *testing.T) {
	Convey("Given a valid job config", t, func() {
		jobConfig := JobConfig{
			SourceID: "/source-id",
			TargetID: "target-id",
			Type:     "static_dataset",
		}

		Convey("When the config is validated", func() {
			errs := jobConfig.Validate()

			Convey("Then no errors should be returend", func() {
				So(errs, ShouldBeNil)
			})
		})
	})

	Convey("Given a job config with a missing parameter", t, func() {
		jobConfig := JobConfig{
			TargetID: "target-id",
			Type:     "static_dataset",
		}

		Convey("When the config is validated", func() {
			errs := jobConfig.Validate()

			Convey("Then an error should be returend", func() {
				So(errs, ShouldHaveLength, 1)
				So(errs, ShouldContain, appErrors.ErrSourceIDNotProvided)
			})
		})
	})

	Convey("Given a job config with a multiple missing parameters", t, func() {
		jobConfig := JobConfig{}

		Convey("When the config is validated", func() {
			errs := jobConfig.Validate()

			Convey("Then errors should be returend", func() {
				So(errs, ShouldHaveLength, 3)
				So(errs, ShouldContain, appErrors.ErrSourceIDNotProvided)
				So(errs, ShouldContain, appErrors.ErrTargetIDNotProvided)
				So(errs, ShouldContain, appErrors.ErrJobTypeNotProvided)
			})
		})
	})

	Convey("Given a job config with an invalid job type", t, func() {
		jobConfig := JobConfig{
			SourceID: "/source-id",
			TargetID: "target-id",
			Type:     "fake job",
		}

		Convey("When the config is validated", func() {
			errs := jobConfig.Validate()

			Convey("Then an error should be returend", func() {
				So(errs, ShouldHaveLength, 1)
				So(errs, ShouldContain, appErrors.ErrJobTypeInvalid)
			})
		})
	})
}

func TestValidateZebedeeURI(t *testing.T) {
	Convey("Given some valid zebedee IDs (URIs)", t, func() {
		validIDs := []string{
			"/economy",
			"/economy/environmentalaccounts/bulletins/greenhousegasintensityprovisionalestimatesuk/2024",
			"/economy/environmentalaccounts/datasets/marineandcoastalmarginsnaturalcapitalaccountsukdetailedsummarytables",
		}
		Convey("When they are validated", func() {
			var errs []error

			for _, id := range validIDs {
				err := validateZebedeeURI(id)
				if err != nil {
					errs = append(errs, err)
				}
			}
			Convey("They should return as valid", func() {
				So(errs, ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given some invalid zebedee IDs (URIs)", t, func() {
		validIDs := []string{
			"/economy?",
			"economy",
			"economy/my-uri",
			"/economy/my-uri#index-this",
			"12087as9c8asc8ca128eu0doasdyasd8y",
		}
		Convey("When they are validated", func() {
			var errs []error

			for _, id := range validIDs {
				err := validateZebedeeURI(id)
				if err != nil {
					errs = append(errs, err)
				}
			}
			Convey("They should return as invalid", func() {
				So(errs, ShouldHaveLength, len(validIDs))
			})
		})
	})
}

func TestValidateDatasetID(t *testing.T) {
	Convey("Given some valid dataset IDs", t, func() {
		validIDs := []string{
			"economy",
			"this-is-a-valid-id",
		}
		Convey("When they are validated", func() {
			var errs []error

			for _, id := range validIDs {
				err := validateDatasetID(id)
				if err != nil {
					errs = append(errs, err)
				}
			}
			Convey("They should return as valid", func() {
				So(errs, ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given some invalid dataset IDs", t, func() {
		validIDs := []string{
			"/economy?",
			"this-is-an-invalid-id-",
			"12087as9c8asc8ca128eu0doasdyasd8y",
		}
		Convey("When they are validated", func() {
			var errs []error

			for _, id := range validIDs {
				err := validateZebedeeURI(id)
				if err != nil {
					errs = append(errs, err)
				}
			}
			Convey("They should return as invalid", func() {
				So(errs, ShouldHaveLength, len(validIDs))
			})
		})
	})
}
