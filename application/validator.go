package application

import (
	"context"
	"errors"
	"net/http"

	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetError "github.com/ONSdigital/dp-dataset-api/apierrors"
	"github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// JobValidator defines the contract for validating job configurations against external systems
type JobValidator interface {
	ValidateSourceID(ctx context.Context, sourceID string, appClients *clients.ClientList) error
	ValidateTargetID(ctx context.Context, targetID string, appClients *clients.ClientList) error
}

var validators = map[domain.JobType]JobValidator{
	domain.JobTypeStaticDataset: &StaticDatasetValidator{},
}

// GetValidator retrieves the appropriate validator for the given jobType
func GetValidator(jobType domain.JobType) (JobValidator, error) {
	validator, exists := validators[jobType]
	if !exists {
		return nil, errors.New("no validator found for jobType")
	}
	return validator, nil
}

type StaticDatasetValidator struct{}

func (v *StaticDatasetValidator) ValidateSourceID(ctx context.Context, sourceID string, appClients *clients.ClientList) error {
	data, err := checkZebedeeURIExists(ctx, appClients.Zebedee, sourceID)
	if err != nil {
		return err
	}

	if data.Type != zebedee.PageTypeDatasetLandingPage {
		return appErrors.ErrSourceIDInvalid
	}

	return nil
}

func (v *StaticDatasetValidator) ValidateTargetID(ctx context.Context, targetID string, appClients *clients.ClientList) error {
	return checkDatasetIDDoesNotExist(ctx, appClients.DatasetAPI, targetID)
}

func checkZebedeeURIExists(ctx context.Context, client clients.ZebedeeClient, uri string) (zebedee.PageData, error) {
	var e zebedee.ErrInvalidZebedeeResponse

	zebedeeData, err := client.GetPageData(ctx, "", "", "en", uri)
	if err != nil {
		if errors.As(err, &e) {
			if e.ActualCode == http.StatusNotFound {
				return zebedee.PageData{}, appErrors.ErrSourceIDInvalid
			}
		}
		log.Error(ctx, "failed to validate source ID with zebedee", err)
		return zebedee.PageData{}, appErrors.ErrSourceIDValidation
	}
	return zebedeeData, nil
}

func checkDatasetIDDoesNotExist(ctx context.Context, client clients.DatasetAPIClient, id string) error {
	_, err := client.GetDataset(ctx, sdk.Headers{}, "", id)
	if err != nil {
		if err == datasetError.ErrDatasetNotFound {
			return nil
		}
		log.Error(ctx, "failed to validate target ID with dataset API", err)
		return appErrors.ErrTargetIDValidation
	}

	return appErrors.ErrTargetIDInvalid
}
