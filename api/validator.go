package api

import (
	"errors"
	"regexp"

	appErrors "github.com/ONSdigital/dis-migration-service/errors"

	"github.com/ONSdigital/dis-migration-service/domain"
)

func validateJobConfig(jc *domain.JobConfig) []error {
	var errs []error

	if jc == nil {
		return []error{appErrors.ErrUnableToParseBody}
	}

	if jc.SourceID == "" {
		errs = append(errs, appErrors.ErrSourceIDNotProvided)
	}

	if jc.TargetID == "" {
		errs = append(errs, appErrors.ErrTargetIDNotProvided)
	}

	err := validateJobType(jc.Type)
	if err != nil {
		errs = append(errs, err)
	}

	//nolint:gocritic //This is a switch statement in anticipation of other types in future.
	switch jc.Type {
	case domain.JobTypeStaticDataset:
		configValidationErrors := validateStaticDatasetJobConfig(jc)
		errs = append(errs, configValidationErrors...)
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func validateJobType(jt domain.JobType) error {
	if jt == "" {
		return appErrors.ErrJobTypeNotProvided
	}

	if !domain.IsValidJobType(jt) {
		return appErrors.ErrInvalidJobType
	}
	return nil
}

func validateStaticDatasetJobConfig(jc *domain.JobConfig) []error {
	var errs []error

	err := validateZebedeeURI(jc.SourceID)
	if err != nil {
		errs = append(errs, appErrors.ErrSourceIDNotValid)
	}

	err = validateDatasetID(jc.TargetID)
	if err != nil {
		errs = append(errs, appErrors.ErrTargetIDNotValid)
	}

	return errs
}

// validateURIPath validates if the given path is a valid URI path
func validateZebedeeURI(path string) error {
	pattern := `^(\/[^\?\/\#\s]+)+$` // Ensures the path starts with '/' and does not contain query strings or hashbangs
	re := regexp.MustCompile(pattern)

	if !re.MatchString(path) {
		return errors.New("URI path must start with '/', not end with '/', not contain query strings or hashbangs")
	}

	// Return nil if the path is valid
	return nil
}

// validateURIPath validates if the given path is a valid URI path
func validateDatasetID(id string) error {
	// Define the regex pattern
	pattern := `^[a-z0-9]+(-[a-z0-9]+)*$`

	// Compile the regex
	re := regexp.MustCompile(pattern)

	// Check if the input matches the pattern
	if !re.MatchString(id) {
		return errors.New("id must be lowercase alphanumeric with optional hyphen separators")
	}

	// Return nil if the input is valid
	return nil
}
