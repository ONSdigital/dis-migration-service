package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID          string     `json:"id" bson:"_id"`
	LastUpdated time.Time  `json:"last_updated" bson:"last_updated"`
	State       JobState   `json:"state" bson:"state"`
	Config      *JobConfig `json:"config" bson:"config"`
	Links       JobLinks   `json:"links" bson:"links"`
}

type JobLinks struct {
	Self *LinkObject `bson:"self,omitempty"       json:"self,omitempty"`
}

func NewJob(cfg *JobConfig, host string) Job {
	id := uuid.New().String()

	links := NewJobLinks(id, host)

	return Job{
		Config:      cfg,
		ID:          id,
		LastUpdated: time.Now().UTC(),
		Links:       links,
		State:       JobStateSubmitted,
	}
}

func NewJobLinks(id, host string) JobLinks {
	return JobLinks{
		Self: &LinkObject{
			HRef: fmt.Sprintf("%s/v1/migration-jobs/%s", host, id),
		},
	}
}
