package domain

import (
	"errors"
	"regexp"

	appErrors "github.com/ONSdigital/dis-migration-service/errors"
)

type JobConfig struct {
	SourceID string  `json:"source_id" bson:"source_id"`
	TargetID string  `json:"target_id" bson:"target_id"`
	Type     JobType `json:"type" bson:"type"`
}

// validateJobConfig only validates whether
func (jc *JobConfig) Validate() []error {
	var errs []error

	if jc == nil {
		return []error{appErrors.ErrUnableToParseBody}
	}

	err := jc.Type.Validate()
	if err != nil {
		errs = append(errs, err)
	}

	if jc.SourceID == "" {
		errs = append(errs, appErrors.ErrSourceIDNotProvided)
	}

	if jc.TargetID == "" {
		errs = append(errs, appErrors.ErrTargetIDNotProvided)
	}

	//nolint:gocritic //This is a switch statement in anticipation of other types in future.
	switch jc.Type {
	case JobTypeStaticDataset:
		configValidationErrors := validateStaticDatasetJobConfig(jc)
		errs = append(errs, configValidationErrors...)
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func validateStaticDatasetJobConfig(jc *JobConfig) []error {
	var errs []error

	if jc.SourceID != "" {
		err := validateZebedeeURI(jc.SourceID)
		if err != nil {
			errs = append(errs, appErrors.ErrSourceIDInvalid)
		}
	}
	if jc.TargetID != "" {
		err := validateDatasetID(jc.TargetID)
		if err != nil {
			errs = append(errs, appErrors.ErrTargetIDInvalid)
		}
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
