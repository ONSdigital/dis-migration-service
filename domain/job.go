package domain

type JobConfig struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Type     string `json:"type"`
}

type Job struct {
	ID          string     `json:"id"`
	LastUpdated string     `json:"last_updated"`
	State       string     `json:"state"`
	Config      *JobConfig `json:"config"`
}

func NewJob(cfg *JobConfig) Job {
	return Job{
		Config: cfg,
		State:  "submitted",
	}
}
