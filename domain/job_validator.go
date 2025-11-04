package domain

import (
	"context"
	"errors"
	"net/http"
	"regexp"

	"github.com/ONSdigital/dis-migration-service/clients"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetErrors "github.com/ONSdigital/dp-dataset-api/apierrors"
	datasetSDK "github.com/ONSdigital/dp-dataset-api/sdk"

	"github.com/ONSdigital/log.go/v2/log"
)

// JobValidator defines the contract for validating job configurations against external systems
//
//go:generate moq -out mock/job_validator.go -pkg mock . JobValidator
type JobValidator interface {
	ValidateSourceID(sourceID string) error
	ValidateSourceIDWithExternal(ctx context.Context, sourceID string, appClients *clients.ClientList) error
	ValidateTargetID(targetID string) error
	ValidateTargetIDWithExternal(ctx context.Context, targetID string, appClients *clients.ClientList) error
}

var validators = map[JobType]JobValidator{
	JobTypeStaticDataset: &StaticDatasetValidator{},
}

// GetValidator retrieves the appropriate validator for the given jobType
func GetValidator(jobType JobType) (JobValidator, error) {
	validator, exists := validators[jobType]
	if !exists {
		return nil, appErrors.ErrJobTypeInvalid
	}
	return validator, nil
}

type StaticDatasetValidator struct{}

func (v *StaticDatasetValidator) ValidateSourceID(sourceID string) error {
	return ValidateZebedeeURI(sourceID)
}

func (v *StaticDatasetValidator) ValidateSourceIDWithExternal(ctx context.Context, sourceID string, appClients *clients.ClientList) error {
	data, err := checkZebedeeURIExists(ctx, appClients.Zebedee, sourceID)
	if err != nil {
		return err
	}

	if data.Type != zebedee.PageTypeDatasetLandingPage {
		return appErrors.ErrSourceIDInvalid
	}

	return nil
}

func (v *StaticDatasetValidator) ValidateTargetID(targetID string) error {
	return ValidateDatasetID(targetID)
}

func (v *StaticDatasetValidator) ValidateTargetIDWithExternal(ctx context.Context, targetID string, appClients *clients.ClientList) error {
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
	_, err := client.GetDataset(ctx, datasetSDK.Headers{}, "", id)
	if err != nil {
		if err.Error() == datasetErrors.ErrDatasetNotFound.Error() {
			return nil
		}
		log.Error(ctx, "failed to validate target ID with dataset API", err)
		return appErrors.ErrTargetIDValidation
	}

	return appErrors.ErrTargetIDInvalid
}

// validateURIPath validates if the given path is a valid URI path
func ValidateZebedeeURI(path string) error {
	pattern := `^(\/[^\?\/\#\s]+)+$` // Ensures the path starts with '/' and does not contain query strings or hashbangs
	re := regexp.MustCompile(pattern)

	if !re.MatchString(path) {
		return appErrors.ErrSourceIDZebedeeURIInvalid
	}

	// Return nil if the path is valid
	return nil
}

// validateURIPath validates if the given path is a valid URI path
func ValidateDatasetID(id string) error {
	// Define the regex pattern
	pattern := `^[a-z0-9]+(-[a-z0-9]+)*$`

	// Compile the regex
	re := regexp.MustCompile(pattern)

	// Check if the input matches the pattern
	if !re.MatchString(id) {
		return appErrors.ErrTargetIDDatasetIDInvalid
	}

	// Return nil if the input is valid
	return nil
}
