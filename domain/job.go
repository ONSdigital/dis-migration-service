package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Job represents a migration job
type Job struct {
	ID          string     `json:"id" bson:"_id"`
	Label       string     `json:"label" bson:"label"`
	LastUpdated time.Time  `json:"last_updated" bson:"last_updated"`
	State       JobState   `json:"state" bson:"state"`
	Config      *JobConfig `json:"config" bson:"config"`
	Links       JobLinks   `json:"links" bson:"links"`
}

// JobLinks contains HATEOS links for a migration job
type JobLinks struct {
	Self   *LinkObject `bson:"self,omitempty"       json:"self,omitempty"`
	Tasks  *LinkObject `bson:"tasks,omitempty"       json:"tasks,omitempty"`
	Events *LinkObject `bson:"events,omitempty"      json:"events,omitempty"`
}

// NewJob creates a new Job instance with the provided configuration
func NewJob(cfg *JobConfig, label string) Job {
	id := uuid.New().String()

	links := NewJobLinks(id)

	return Job{
		Config:      cfg,
		ID:          id,
		Label:       label,
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
		Tasks: &LinkObject{
			HRef: fmt.Sprintf("/v1/migration-jobs/%s/tasks", id),
		},
		Events: &LinkObject{
			HRef: fmt.Sprintf("/v1/migration-jobs/%s/events", id),
		},
	}
}
