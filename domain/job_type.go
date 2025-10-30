package domain

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
