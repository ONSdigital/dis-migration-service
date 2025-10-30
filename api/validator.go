package api

import (
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

	if jc.Type == "" {
		errs = append(errs, appErrors.ErrJobTypeNotProvided)
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
