package domain

import (
	"fmt"
)

// Event represents a migration job event (state change)
type Event struct {
	ID          string     `json:"id" bson:"_id"`
	CreatedAt   string     `json:"created_at" bson:"created_at"`
	RequestedBy *User      `json:"requested_by" bson:"requested_by"`
	Action      string     `json:"action" bson:"action"`
	JobID       string     `json:"job_id" bson:"job_id"`
	Links       EventLinks `json:"links" bson:"links"`
}

// User represents a user who initiated an action
type User struct {
	ID    string `json:"id" bson:"id"`
	Email string `json:"email,omitempty" bson:"email,omitempty"`
}

// EventLinks contains HATEOAS links for a migration event
type EventLinks struct {
	Self *LinkObject `bson:"self,omitempty" json:"self,omitempty"`
	Job  *LinkObject `bson:"job,omitempty" json:"job,omitempty"`
}

// NewEventLinks creates EventLinks for an event with the given ID and jobID
func NewEventLinks(id, jobID, host string) EventLinks {
	return EventLinks{
		Self: &LinkObject{
			HRef: fmt.Sprintf("%s/v1/migration-jobs/%s/events/%s", host, jobID, id),
		},
		Job: &LinkObject{
			HRef: fmt.Sprintf("%s/v1/migration-jobs/%s", host, jobID),
		},
	}
}
