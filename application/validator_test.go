package application

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/ONSdigital/dis-migration-service/clients"
	clientMocks "github.com/ONSdigital/dis-migration-service/clients/mock"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-dataset-api/sdk"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	zebedeeErrorPath    = "/error"
	zebedeeNotFoundPath = "/not-found"
	zebedeeValidPath    = "/found"

	datasetErrorID    = "error"
	datasetNotFoundID = "not-found"
	datasetValidID    = "found"
)

func TestStaticDatasetValidator(t *testing.T) {
	zebedeeMock := &clientMocks.ZebedeeClientMock{
		GetPageDataFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedee.PageData, error) {
			switch path {
			case zebedeeErrorPath:
				return zebedee.PageData{}, errors.New("unexpected error")
			case zebedeeValidPath:
				return zebedee.PageData{}, nil
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
				return datasetModels.Dataset{}, &clientMocks.ClientErrorMock{
					CodeFunc: func() int { return http.StatusNotFound },
				}
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
		validator := StaticDatasetValidator{}

		Convey("When the source is validated", func() {
			err := validator.ValidateSourceID(ctx, zebedeeValidPath, &mockClientlist)

			Convey("Then no error should be returend", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a zebedee source ID that returns an unexpected error", t, func() {
		validator := StaticDatasetValidator{}

		Convey("When the source is validated", func() {
			err := validator.ValidateSourceID(ctx, zebedeeErrorPath, &mockClientlist)

			Convey("Then an error should be returend", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a zebedee source ID that returns not found error", t, func() {
		validator := StaticDatasetValidator{}

		Convey("When the source is validated", func() {
			err := validator.ValidateSourceID(ctx, zebedeeNotFoundPath, &mockClientlist)

			Convey("Then an error should be returend", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a dataset target ID that returns as not found", t, func() {
		validator := StaticDatasetValidator{}

		Convey("When the target ID is validated", func() {
			err := validator.ValidateTargetID(ctx, datasetNotFoundID, &mockClientlist)

			Convey("Then no error should be returend", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a dataset target ID that returns a value", t, func() {
		validator := StaticDatasetValidator{}

		Convey("When the target is validated", func() {
			err := validator.ValidateTargetID(ctx, datasetValidID, &mockClientlist)

			Convey("Then an error should be returend", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a dataset target ID that returns an unexpected error", t, func() {
		validator := StaticDatasetValidator{}

		Convey("When the target ID is validated", func() {
			err := validator.ValidateTargetID(ctx, datasetErrorID, &mockClientlist)

			Convey("Then an error should be returend", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

}
