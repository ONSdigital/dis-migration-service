package domain

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/ONSdigital/dis-migration-service/clients"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetErrors "github.com/ONSdigital/dp-dataset-api/apierrors"
	datasetSDK "github.com/ONSdigital/dp-dataset-api/sdk"

	"github.com/ONSdigital/log.go/v2/log"
)

// JobValidator defines the contract for validating job configurations
// against external systems
//
//go:generate moq -out mock/job_validator.go -pkg mock . JobValidator
type JobValidator interface {
	ValidateSourceID(sourceID string) error
	ValidateSourceIDWithExternal(ctx context.Context, sourceID string, appClients *clients.ClientList, userAuthToken string) (string, error)
	ValidateTargetID(targetID string) error
	ValidateTargetIDWithExternal(ctx context.Context, targetID string, appClients *clients.ClientList, userAuthToken string) error
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

// StaticDatasetValidator implements JobValidator for static dataset
// migration jobs
type StaticDatasetValidator struct{}

// ValidateSourceID validates if the given source ID is a valid
// Zebedee URI
func (v *StaticDatasetValidator) ValidateSourceID(sourceID string) error {
	return ValidateZebedeeURI(sourceID)
}

// ValidateSourceIDWithExternal validates if the given source ID exists in
// Zebedee and is of the correct type
func (v *StaticDatasetValidator) ValidateSourceIDWithExternal(ctx context.Context, sourceID string, appClients *clients.ClientList, userAuthToken string) (string, error) {
	data, err := checkZebedeeURIExists(ctx, appClients.Zebedee, sourceID, userAuthToken)
	if err != nil {
		return "", err
	}

	if data.Type != zebedee.PageTypeDatasetLandingPage {
		log.Error(ctx, data.Type, appErrors.ErrSourceIDInvalid)
		return "", appErrors.ErrSourceIDInvalid
	}

	// Extract and validate title
	title := strings.TrimSpace(data.Description.Title)
	if title == "" {
		return "", appErrors.ErrSourceTitleNotFound
	}

	return title, nil
}

// ValidateTargetID validates if the given id is a valid dataset ID
func (v *StaticDatasetValidator) ValidateTargetID(targetID string) error {
	return ValidateDatasetID(targetID)
}

// ValidateTargetIDWithExternal validates that the target dataset
// ID does not already exist
func (v *StaticDatasetValidator) ValidateTargetIDWithExternal(ctx context.Context, targetID string, appClients *clients.ClientList, userAuthToken string) error {
	return checkDatasetIDDoesNotExist(ctx, appClients.DatasetAPI, targetID, userAuthToken)
}

func checkZebedeeURIExists(ctx context.Context, client clients.ZebedeeClient, uri, userAuthToken string) (zebedee.PageData, error) {
	var e zebedee.ErrInvalidZebedeeResponse
	zebedeeData, err := client.GetPageData(ctx, userAuthToken, "", "en", uri)
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

func checkDatasetIDDoesNotExist(ctx context.Context, client datasetSDK.Clienter, id, userAuthToken string) error {
	_, err := client.GetDataset(ctx, datasetSDK.Headers{AccessToken: userAuthToken}, id)
	if err != nil {
		if err.Error() == datasetErrors.ErrDatasetNotFound.Error() {
			return nil
		}
		log.Error(ctx, "failed to validate target ID with dataset API", err)
		return appErrors.ErrTargetIDValidation
	}

	return appErrors.ErrTargetIDInvalid
}

// ValidateZebedeeURI validates if the given path is a valid URI
// path
func ValidateZebedeeURI(path string) error {
	pattern := `^(\/[^\?\/\#\s]+)+$` // Ensures the path starts with '/' and does not contain query strings or hashbangs
	re := regexp.MustCompile(pattern)

	if !re.MatchString(path) {
		return appErrors.ErrSourceIDZebedeeURIInvalid
	}

	// Return nil if the path is valid
	return nil
}

// ValidateDatasetID validates if the given id is a valid dataset ID
func ValidateDatasetID(id string) error {
	// Check length
	if len(id) < 1 || len(id) > 100 {
		return appErrors.ErrTargetIDDatasetIDInvalid
	}

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
