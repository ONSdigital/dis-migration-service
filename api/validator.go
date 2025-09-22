package api

import (
	apiErrors "github.com/ONSdigital/dis-migration-service/api/errors"

	"github.com/ONSdigital/dis-migration-service/domain"
)

func validateJobConfig(jc *domain.JobConfig) []apiErrors.APIError {
	var errs []apiErrors.APIError

	if jc == nil {
		return []apiErrors.APIError{apiErrors.ErrUnableToParseBody}
	}

	if jc.SourceID == "" {
		errs = append(errs, apiErrors.ErrSourceIDNotProvided)
	}

	if jc.TargetID == "" {
		errs = append(errs, apiErrors.ErrTargetIDNotProvided)
	}

	if jc.Type == "" {
		errs = append(errs, apiErrors.ErrJobTypeNotProvided)
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
