package application

import (
	"context"
	"errors"
	"net/http"

	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
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
	return checkZebedeeURIExists(ctx, appClients.Zebedee, sourceID)
}

func (v *StaticDatasetValidator) ValidateTargetID(ctx context.Context, targetID string, appClients *clients.ClientList) error {
	return checkDatasetIDDoesNotExist(ctx, appClients.DatasetAPI, targetID)
}

func checkZebedeeURIExists(ctx context.Context, client clients.ZebedeeClient, uri string) error {
	var e zebedee.ErrInvalidZebedeeResponse

	_, err := client.GetPageData(ctx, "", "", "en", uri)
	if err != nil {
		if errors.As(err, &e) {
			if e.ActualCode == http.StatusNotFound {
				return appErrors.ErrSourceIDInvalid
			}
		}
		log.Error(ctx, "failed to validate source ID with zebedee", err)
		return appErrors.ErrSourceIDValidation
	}
	return nil
}

func checkDatasetIDDoesNotExist(ctx context.Context, client clients.DatasetAPIClient, id string) error {
	_, err := client.GetDataset(ctx, sdk.Headers{}, "", id)
	if clientErr, ok := err.(clients.ClientError); ok {
		if clientErr.Code() == http.StatusNotFound {
			return nil
		} else {
			log.Error(ctx, "failed to validate target ID with dataset API", err)
			return appErrors.ErrTargetIDValidation
		}
	}
	return appErrors.ErrTargetIDInvalid
}
