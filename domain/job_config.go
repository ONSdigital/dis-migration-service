package domain

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/clients"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
)

// JobConfig represents the configuration for a migration job
type JobConfig struct {
	SourceID  string       `json:"source_id" bson:"source_id"`
	TargetID  string       `json:"target_id" bson:"target_id"`
	Type      JobType      `json:"type" bson:"type"`
	Validator JobValidator `json:"-" bson:"-"`
}

// ValidateInternal performs internal validation of the JobConfig fields
func (jc *JobConfig) ValidateInternal() []error {
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

	err := jc.Type.Validate()
	if err != nil {
		errs = append(errs, err)
	} else if jc.Validator == nil {
		jc.Validator, err = GetValidator(jc.Type)
		if err != nil {
			errs = append(errs, appErrors.ErrJobTypeInvalid)
		}
	}

	if len(errs) > 0 {
		return errs
	}

	if jc.Validator != nil {
		err = jc.Validator.ValidateSourceID(jc.SourceID)
		if err != nil {
			errs = append(errs, err)
		}
		err = jc.Validator.ValidateTargetID(jc.TargetID)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// ValidateExternal performs validation of the JobConfig fields
// against external systems
func (jc *JobConfig) ValidateExternal(ctx context.Context, appClients clients.ClientList) (string, error) {
	var err error

	if jc.Validator == nil {
		jc.Validator, err = GetValidator(jc.Type)
		if err != nil {
			return "", err
		}
	}

	title, err := jc.Validator.ValidateSourceIDWithExternal(ctx, jc.SourceID, &appClients)
	if err != nil {
		return "", err
	}

	err = jc.Validator.ValidateTargetIDWithExternal(ctx, jc.TargetID, &appClients)
	if err != nil {
		return "", err
	}

	return title, nil
}
