package domain

import appErrors "github.com/ONSdigital/dis-migration-service/errors"

type JobType string

const (
	JobTypeStaticDataset JobType = "static_dataset"
)

func IsValidJobType(state JobType) bool {
	switch state {
	case JobTypeStaticDataset:
		return true
	default:
		return false
	}
}

func (jt JobType) Validate() error {
	if jt == "" {
		return appErrors.ErrJobTypeNotProvided
	}

	if !IsValidJobType(jt) {
		return appErrors.ErrJobTypeInvalid
	}
	return nil
}
