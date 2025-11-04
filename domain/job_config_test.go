package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	mock "github.com/ONSdigital/dis-migration-service/domain/mock"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateJobConfigInternal(t *testing.T) {
	Convey("Given a valid job config with a mock validator that returns ok", t, func() {
		mockValidator := &mock.JobValidatorMock{
			ValidateSourceIDFunc: func(sourceID string) error {
				return nil
			},
			ValidateTargetIDFunc: func(targetID string) error {
				return nil
			},
		}

		jobConfig := domain.JobConfig{
			SourceID:  "/source-id",
			TargetID:  "target-id",
			Type:      domain.JobTypeStaticDataset,
			Validator: mockValidator,
		}

		Convey("When the config is validated", func() {
			errs := jobConfig.ValidateInternal()

			Convey("Then no errors should be returend", func() {
				So(errs, ShouldBeNil)

				Convey("And the validator should have be called for internal validation", func() {
					So(len(mockValidator.ValidateSourceIDCalls()), ShouldEqual, 1)
					So(len(mockValidator.ValidateTargetIDCalls()), ShouldEqual, 1)
				})
			})
		})
	})

	Convey("Given a valid job config with a mock validator that returns errors", t, func() {
		testErrorSourceID := errors.New("this is a test error for source id")
		testErrorTargetID := errors.New("this is a test error for target id")

		mockValidator := &mock.JobValidatorMock{
			ValidateSourceIDFunc: func(sourceID string) error {
				return testErrorSourceID
			},
			ValidateTargetIDFunc: func(targetID string) error {
				return testErrorTargetID
			},
		}

		jobConfig := domain.JobConfig{
			SourceID:  "/source-id",
			TargetID:  "target-id",
			Type:      domain.JobTypeStaticDataset,
			Validator: mockValidator,
		}

		Convey("When the config is validated", func() {
			errs := jobConfig.ValidateInternal()

			Convey("Then errors should be returned", func() {
				So(errs, ShouldContain, testErrorSourceID)
				So(errs, ShouldContain, testErrorTargetID)
			})
		})
	})

	Convey("Given a job config with a missing parameter", t, func() {
		mockValidator := &mock.JobValidatorMock{
			ValidateSourceIDFunc: func(sourceID string) error {
				return nil
			},
			ValidateTargetIDFunc: func(targetID string) error {
				return nil
			},
		}

		jobConfig := domain.JobConfig{
			TargetID: "target-id",
			Type:     domain.JobTypeStaticDataset,
		}

		Convey("When the config is validated", func() {
			errs := jobConfig.ValidateInternal()

			Convey("Then an error should be returend", func() {
				So(errs, ShouldHaveLength, 1)
				So(errs, ShouldContain, appErrors.ErrSourceIDNotProvided)

				Convey("And the validator should not be called", func() {
					So(len(mockValidator.ValidateSourceIDCalls()), ShouldEqual, 0)
					So(len(mockValidator.ValidateTargetIDCalls()), ShouldEqual, 0)
				})
			})
		})
	})

	Convey("Given a job config with a multiple missing parameters", t, func() {
		jobConfig := domain.JobConfig{}

		Convey("When the config is validated", func() {
			errs := jobConfig.ValidateInternal()

			Convey("Then errors should be returend", func() {
				So(errs, ShouldHaveLength, 3)
				So(errs, ShouldContain, appErrors.ErrSourceIDNotProvided)
				So(errs, ShouldContain, appErrors.ErrTargetIDNotProvided)
				So(errs, ShouldContain, appErrors.ErrJobTypeNotProvided)
			})
		})
	})

	Convey("Given a job config with an invalid job type", t, func() {
		mockValidator := &mock.JobValidatorMock{
			ValidateSourceIDFunc: func(sourceID string) error {
				return nil
			},
			ValidateTargetIDFunc: func(targetID string) error {
				return nil
			},
		}

		jobConfig := domain.JobConfig{
			SourceID:  "/source-id",
			TargetID:  "target-id",
			Type:      "fake job",
			Validator: mockValidator,
		}

		Convey("When the config is validated", func() {
			errs := jobConfig.ValidateInternal()

			Convey("Then an error should be returend", func() {
				So(errs, ShouldHaveLength, 1)
				So(errs, ShouldContain, appErrors.ErrJobTypeInvalid)

				Convey("And the validator should not be called", func() {
					So(len(mockValidator.ValidateSourceIDCalls()), ShouldEqual, 0)
					So(len(mockValidator.ValidateTargetIDCalls()), ShouldEqual, 0)
				})
			})
		})
	})
}

func TestValidateJobConfigExternal(t *testing.T) {
	Convey("Given a valid job config with a mock validator that returns ok", t, func() {
		mockValidator := &mock.JobValidatorMock{
			ValidateSourceIDWithExternalFunc: func(ctx context.Context, sourceID string, appClients *clients.ClientList) error {
				return nil
			},
			ValidateTargetIDWithExternalFunc: func(ctx context.Context, targetID string, appClients *clients.ClientList) error {
				return nil
			},
		}

		jobConfig := domain.JobConfig{
			SourceID:  "/source-id",
			TargetID:  "target-id",
			Type:      domain.JobTypeStaticDataset,
			Validator: mockValidator,
		}

		ctx := context.Background()
		mockClients := clients.ClientList{}

		Convey("When the config is validated", func() {
			errs := jobConfig.ValidateExternal(ctx, mockClients)

			Convey("Then no errors should be returend", func() {
				So(errs, ShouldBeNil)

				Convey("And the validator should have be called for internal validation", func() {
					So(len(mockValidator.ValidateSourceIDWithExternalCalls()), ShouldEqual, 1)
					So(len(mockValidator.ValidateTargetIDWithExternalCalls()), ShouldEqual, 1)
				})
			})
		})
	})

	Convey("Given a valid job config with a mock validator that returns an error for the sourceID", t, func() {
		testErrorSourceID := errors.New("this is a test error for source id")

		mockValidator := &mock.JobValidatorMock{
			ValidateSourceIDWithExternalFunc: func(ctx context.Context, sourceID string, appClients *clients.ClientList) error {
				return testErrorSourceID
			},
			ValidateTargetIDWithExternalFunc: func(ctx context.Context, targetID string, appClients *clients.ClientList) error {
				return nil
			},
		}

		jobConfig := domain.JobConfig{
			SourceID:  "/source-id",
			TargetID:  "target-id",
			Type:      domain.JobTypeStaticDataset,
			Validator: mockValidator,
		}

		ctx := context.Background()
		mockClients := clients.ClientList{}

		Convey("When the config is validated", func() {
			err := jobConfig.ValidateExternal(ctx, mockClients)

			Convey("Then errors should be returned", func() {
				So(err, ShouldEqual, testErrorSourceID)
			})
		})
	})

	Convey("Given a valid job config with a mock validator that returns an error for the targetID", t, func() {
		testErrorTargetID := errors.New("this is a test error for source id")

		mockValidator := &mock.JobValidatorMock{
			ValidateSourceIDWithExternalFunc: func(ctx context.Context, sourceID string, appClients *clients.ClientList) error {
				return nil
			},
			ValidateTargetIDWithExternalFunc: func(ctx context.Context, targetID string, appClients *clients.ClientList) error {
				return testErrorTargetID
			},
		}

		jobConfig := domain.JobConfig{
			SourceID:  "/source-id",
			TargetID:  "target-id",
			Type:      domain.JobTypeStaticDataset,
			Validator: mockValidator,
		}

		ctx := context.Background()
		mockClients := clients.ClientList{}

		Convey("When the config is validated", func() {
			err := jobConfig.ValidateExternal(ctx, mockClients)

			Convey("Then errors should be returned", func() {
				So(err, ShouldEqual, testErrorTargetID)
			})
		})
	})

	Convey("Given a job config with an invalid type", t, func() {
		jobConfig := domain.JobConfig{
			SourceID: "/source-id",
			TargetID: "target-id",
			Type:     "invalid type",
		}

		ctx := context.Background()
		mockClients := clients.ClientList{}

		Convey("When the config is validated", func() {
			err := jobConfig.ValidateExternal(ctx, mockClients)

			Convey("Then an error should be returned", func() {
				So(err, ShouldEqual, appErrors.ErrJobTypeInvalid)
			})
		})
	})
}
