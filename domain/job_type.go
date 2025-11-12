package domain

import appErrors "github.com/ONSdigital/dis-migration-service/errors"

// JobType represents the type of migration job
type JobType string

//nolint:godoclint // Documentation for these constants is provided in the JobType type comment.
const (
	JobTypeStaticDataset JobType = "static_dataset"
)

// IsValidJobType checks if the provided JobType is valid
func IsValidJobType(state JobType) bool {
	switch state {
	case JobTypeStaticDataset:
		return true
	default:
		return false
	}
}

// Validate checks if the JobType is valid and not empty
func (jt JobType) Validate() error {
	if jt == "" {
		return appErrors.ErrJobTypeNotProvided
	}

	if !IsValidJobType(jt) {
		return appErrors.ErrJobTypeInvalid
	}
	return nil
}
