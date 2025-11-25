package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Job represents a migration job
type Job struct {
	ID          string     `json:"id" bson:"_id"`
	LastUpdated time.Time  `json:"last_updated" bson:"last_updated"`
	State       JobState   `json:"state" bson:"state"`
	Config      *JobConfig `json:"config" bson:"config"`
	Links       JobLinks   `json:"links" bson:"links"`
}

// JobLinks contains HATEOS links for a migration job
type JobLinks struct {
	Self *LinkObject `bson:"self,omitempty"       json:"self,omitempty"`
}

// NewJob creates a new Job instance with the provided configuration
func NewJob(cfg *JobConfig) Job {
	id := uuid.New().String()

	links := NewJobLinks(id)

	return Job{
		Config:      cfg,
		ID:          id,
		LastUpdated: time.Now().UTC(),
		Links:       links,
		State:       JobStateSubmitted,
	}
}

// NewJobLinks creates JobLinks for a job with the given ID
func NewJobLinks(id string) JobLinks {
	return JobLinks{
		Self: &LinkObject{
			HRef: fmt.Sprintf("/v1/migration-jobs/%s", id),
		},
	}
}
