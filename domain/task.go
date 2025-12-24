package domain

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Task represents a migration task
type Task struct {
	ID          string        `json:"id" bson:"_id"`
	JobNumber   int           `json:"job_number" bson:"job_number"`
	LastUpdated time.Time     `json:"last_updated" bson:"last_updated"`
	Source      *TaskMetadata `json:"source" bson:"source"`
	State       TaskState     `json:"state" bson:"state"`
	Target      *TaskMetadata `json:"target" bson:"target"`
	Type        TaskType      `json:"type" bson:"type"`
	Links       TaskLinks     `json:"links" bson:"links"`
}

// NewTask creates a new Task instance with the provided configuration
func NewTask(jobNumber int) Task {
	id := uuid.New().String()

	links := NewTaskLinks(id, strconv.Itoa(jobNumber))

	return Task{
		ID:          id,
		JobNumber:   jobNumber,
		LastUpdated: time.Now().UTC(),
		Links:       links,
		State:       TaskStateSubmitted,
	}
}

// TaskMetadata represents metadata about a task's source or target
type TaskMetadata struct {
	ID        string `json:"id" bson:"id"`
	DatasetID string `json:"dataset_id,omitempty" bson:"dataset_id"`
	Label     string `json:"label" bson:"label"`
}

// TaskType represents the type of migration task
type TaskType string

const (
	// TaskTypeDatasetSeries indicates a dataset series task
	TaskTypeDatasetSeries TaskType = "dataset_series"
	// TaskTypeDatasetEdition indicates a dataset edition task
	TaskTypeDatasetEdition TaskType = "dataset_edition"
	// TaskTypeDatasetVersion indicates a dataset version task
	TaskTypeDatasetVersion TaskType = "dataset_version"
	// TaskTypeDatasetDownload indicates a dataset download task
	TaskTypeDatasetDownload TaskType = "dataset_download"
)

// TaskLinks contains HATEOAS links for a migration task
type TaskLinks struct {
	Self *LinkObject `bson:"self,omitempty" json:"self,omitempty"`
	Job  *LinkObject `bson:"job,omitempty" json:"job,omitempty"`
}

// NewTaskLinks creates TaskLinks for a task
func NewTaskLinks(id, jobNumber string) TaskLinks {
	return TaskLinks{
		Self: &LinkObject{
			HRef: fmt.Sprintf("/v1/migration-jobs/%s/tasks/%s", jobNumber, id),
		},
		Job: &LinkObject{
			HRef: fmt.Sprintf("/v1/migration-jobs/%s", jobNumber),
		},
	}
}
