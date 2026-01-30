package domain_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/ONSdigital/dis-migration-service/clients"
	clientMocks "github.com/ONSdigital/dis-migration-service/clients/mock"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetError "github.com/ONSdigital/dp-dataset-api/apierrors"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetSDK "github.com/ONSdigital/dp-dataset-api/sdk"
	datasetMocks "github.com/ONSdigital/dp-dataset-api/sdk/mocks"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	testTitle = "Test Dataset Title"

	zebedeeErrorPath      = "/error"
	zebedeeNotFoundPath   = "/not-found"
	zebedeeWrongType      = "/wrong-type"
	zebedeeValidPath      = "/found"
	zebedeeEmptyTitlePath = "/empty-title"

	datasetErrorID    = "error"
	datasetNotFoundID = "not-found"
	datasetValidID    = "found"
)

func TestStaticDatasetValidatorWithExternal(t *testing.T) {
	zebedeeMock := &clientMocks.ZebedeeClientMock{
		GetPageDataFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedee.PageData, error) {
			switch path {
			case zebedeeErrorPath:
				return zebedee.PageData{}, errors.New("unexpected error")
			case zebedeeValidPath:
				return zebedee.PageData{
					Type: zebedee.PageTypeDatasetLandingPage,
					Description: zebedee.Description{
						Title: testTitle,
					},
				}, nil
			case zebedeeWrongType:
				return zebedee.PageData{
					Type: zebedee.PageTypeBulletin,
					Description: zebedee.Description{
						Title: testTitle,
					},
				}, nil
			case zebedeeNotFoundPath:
				return zebedee.PageData{}, zebedee.ErrInvalidZebedeeResponse{ActualCode: http.StatusNotFound}
			case zebedeeEmptyTitlePath:
				return zebedee.PageData{
					Type: zebedee.PageTypeDatasetLandingPage,
					Description: zebedee.Description{
						Title: "",
					},
				}, nil
			}
			return zebedee.PageData{}, errors.New("unexpected mock path")
		},
	}

	datasetAPIMock := &datasetMocks.ClienterMock{
		GetDatasetFunc: func(ctx context.Context, headers datasetSDK.Headers, datasetID string) (datasetModels.Dataset, error) {
			switch datasetID {
			case datasetErrorID:
				return datasetModels.Dataset{}, errors.New("unexpected error")
			case datasetValidID:
				return datasetModels.Dataset{}, nil
			case datasetNotFoundID:
				return datasetModels.Dataset{}, datasetError.ErrDatasetNotFound
			}
			return datasetModels.Dataset{}, errors.New("unexpected mock id")
		},
	}

	mockClientlist := clients.ClientList{
		Zebedee:    zebedeeMock,
		DatasetAPI: datasetAPIMock,
	}

	ctx := context.Background()

	Convey("Given a valid zebedee source ID with a title", t, func() {
		validator := domain.StaticDatasetValidator{}

		Convey("When the source is validated", func() {
			title, err := validator.ValidateSourceIDWithExternal(ctx, zebedeeValidPath, &mockClientlist, testUserAuthToken)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)

				Convey("And the title should be returned", func() {
					So(title, ShouldEqual, testTitle)
				})
			})
		})
	})

	Convey("Given a zebedee source ID with an empty title", t, func() {
		validator := domain.StaticDatasetValidator{}

		Convey("When the source is validated", func() {
			title, err := validator.ValidateSourceIDWithExternal(ctx, zebedeeEmptyTitlePath, &mockClientlist, testUserAuthToken)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, appErrors.ErrSourceTitleNotFound)

				Convey("And the title should be empty", func() {
					So(title, ShouldEqual, "")
				})
			})
		})
	})

	Convey("Given a zebedee source ID that returns an unexpected error", t, func() {
		validator := domain.StaticDatasetValidator{}

		Convey("When the source is validated", func() {
			title, err := validator.ValidateSourceIDWithExternal(ctx, zebedeeErrorPath, &mockClientlist, testUserAuthToken)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the title should be empty", func() {
					So(title, ShouldEqual, "")
				})
			})
		})
	})

	Convey("Given a zebedee source ID that returns not found error", t, func() {
		validator := domain.StaticDatasetValidator{}

		Convey("When the source is validated", func() {
			title, err := validator.ValidateSourceIDWithExternal(ctx, zebedeeNotFoundPath, &mockClientlist, testUserAuthToken)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the title should be empty", func() {
					So(title, ShouldEqual, "")
				})
			})
		})
	})

	Convey("Given a zebedee source ID that returns the wrong type", t, func() {
		validator := domain.StaticDatasetValidator{}

		Convey("When the source is validated", func() {
			title, err := validator.ValidateSourceIDWithExternal(ctx, zebedeeWrongType, &mockClientlist, testUserAuthToken)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the title should be empty", func() {
					So(title, ShouldEqual, "")
				})
			})
		})
	})

	Convey("Given a dataset target ID that returns as not found", t, func() {
		validator := domain.StaticDatasetValidator{}

		Convey("When the target ID is validated", func() {
			err := validator.ValidateTargetIDWithExternal(ctx, datasetNotFoundID, &mockClientlist, testUserAuthToken)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a dataset target ID that returns a value", t, func() {
		validator := domain.StaticDatasetValidator{}

		Convey("When the target is validated", func() {
			err := validator.ValidateTargetIDWithExternal(ctx, datasetValidID, &mockClientlist, testUserAuthToken)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a dataset target ID that returns an unexpected error", t, func() {
		validator := domain.StaticDatasetValidator{}

		Convey("When the target ID is validated", func() {
			err := validator.ValidateTargetIDWithExternal(ctx, datasetErrorID, &mockClientlist, testUserAuthToken)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
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
				err := domain.ValidateZebedeeURI(id)
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
				err := domain.ValidateZebedeeURI(id)
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
				err := domain.ValidateDatasetID(id)
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
			"a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6A7B8C9D0E1F2G3H4I5J6K7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2", // 101 character string
			"",
		}
		Convey("When they are validated", func() {
			var errs []error

			for _, id := range validIDs {
				err := domain.ValidateZebedeeURI(id)
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
