package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetErrors "github.com/ONSdigital/dp-dataset-api/apierrors"
	"github.com/ONSdigital/dp-dataset-api/sdk"
)

// JobValidator defines the contract for validating job configurations
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
	_, err := client.GetPageData(ctx, "", "", "en", uri)
	var e zebedee.ErrInvalidZebedeeResponse

	if err != nil {
		if errors.As(err, &e) {
			if e.ActualCode == http.StatusNotFound {
				return ErrSourceIDNotFound
			}
		}
		return fmt.Errorf("%w: %v", ErrSourceIDValidation, err)
	}
	return nil
}

func checkDatasetIDDoesNotExist(ctx context.Context, client clients.DatasetAPIClient, id string) error {
	_, err := client.GetDataset(ctx, sdk.Headers{}, "", id)
	if err != nil {
		if err == datasetErrors.ErrDatasetNotFound {
			return ErrTargetIDFound
		}
		return fmt.Errorf("%w: %v", ErrTargetIDValidation, err)
	}
	return nil
}
