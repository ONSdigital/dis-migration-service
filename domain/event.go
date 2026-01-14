package domain

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const (
	// SystemUserID is the default user ID used for events when
	// no user ID is provided. This is typically used for
	// automated or system-generated events.
	SystemUserID = "system"
)

// Event represents a migration job event (state change)
type Event struct {
	ID          string     `json:"id" bson:"_id"`
	CreatedAt   string     `json:"created_at" bson:"created_at"`
	RequestedBy *User      `json:"requested_by" bson:"requested_by"`
	Action      string     `json:"action" bson:"action"`
	JobNumber   int        `json:"job_number" bson:"job_number"`
	Links       EventLinks `json:"links" bson:"links"`
}

// NewEvent creates a new Event with the provided parameters
func NewEvent(jobNumber int, action, userID string) *Event {
	// Use SystemUserID as default if no user ID provided
	if userID == "" {
		userID = SystemUserID
	}

	return &Event{
		ID:        uuid.New().String(),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		RequestedBy: &User{
			ID: userID,
		},
		Action:    action,
		JobNumber: jobNumber,
		Links:     NewEventLinks(uuid.New().String(), strconv.Itoa(jobNumber)),
	}
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
func NewEventLinks(id, jobNumber string) EventLinks {
	return EventLinks{
		Self: &LinkObject{
			HRef: fmt.Sprintf("/v1/migration-jobs/%s/events/%s", jobNumber, id),
		},
		Job: &LinkObject{
			HRef: fmt.Sprintf("/v1/migration-jobs/%s", jobNumber),
		},
	}
}
