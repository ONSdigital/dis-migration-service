package domain_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/ONSdigital/dis-migration-service/clients"
	clientMocks "github.com/ONSdigital/dis-migration-service/clients/mock"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetError "github.com/ONSdigital/dp-dataset-api/apierrors"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-dataset-api/sdk"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	zebedeeErrorPath    = "/error"
	zebedeeNotFoundPath = "/not-found"
	zebedeeWrongType    = "/wrong-type"
	zebedeeValidPath    = "/found"

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
				return zebedee.PageData{Type: zebedee.PageTypeDatasetLandingPage}, nil
			case zebedeeWrongType:
				return zebedee.PageData{Type: zebedee.PageTypeBulletin}, nil
			case zebedeeNotFoundPath:
				return zebedee.PageData{}, zebedee.ErrInvalidZebedeeResponse{ActualCode: http.StatusNotFound}
			}
			return zebedee.PageData{}, errors.New("unexpected mock path")
		},
	}

	datasetAPIMock := &clientMocks.DatasetAPIClientMock{
		GetDatasetFunc: func(ctx context.Context, headers sdk.Headers, collectionID, datasetID string) (datasetModels.Dataset, error) {
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

	Convey("Given a valid zebedee source ID", t, func() {
		validator := domain.StaticDatasetValidator{}

		Convey("When the source is validated", func() {
			err := validator.ValidateSourceIDWithExternal(ctx, zebedeeValidPath, &mockClientlist)
			Convey("Then no error should be returend", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a zebedee source ID that returns an unexpected error", t, func() {
		validator := domain.StaticDatasetValidator{}

		Convey("When the source is validated", func() {
			err := validator.ValidateSourceIDWithExternal(ctx, zebedeeErrorPath, &mockClientlist)

			Convey("Then an error should be returend", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a zebedee source ID that returns not found error", t, func() {
		validator := domain.StaticDatasetValidator{}

		Convey("When the source is validated", func() {
			err := validator.ValidateSourceIDWithExternal(ctx, zebedeeNotFoundPath, &mockClientlist)

			Convey("Then an error should be returend", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a zebedee source ID that returns the wrong type", t, func() {
		validator := domain.StaticDatasetValidator{}

		Convey("When the source is validated", func() {
			err := validator.ValidateSourceIDWithExternal(ctx, zebedeeWrongType, &mockClientlist)

			Convey("Then an error should be returend", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a dataset target ID that returns as not found", t, func() {
		validator := domain.StaticDatasetValidator{}

		Convey("When the target ID is validated", func() {
			err := validator.ValidateTargetIDWithExternal(ctx, datasetNotFoundID, &mockClientlist)

			Convey("Then no error should be returend", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a dataset target ID that returns a value", t, func() {
		validator := domain.StaticDatasetValidator{}

		Convey("When the target is validated", func() {
			err := validator.ValidateTargetIDWithExternal(ctx, datasetValidID, &mockClientlist)

			Convey("Then an error should be returend", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a dataset target ID that returns an unexpected error", t, func() {
		validator := domain.StaticDatasetValidator{}

		Convey("When the target ID is validated", func() {
			err := validator.ValidateTargetIDWithExternal(ctx, datasetErrorID, &mockClientlist)

			Convey("Then an error should be returend", func() {
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
