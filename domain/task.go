package domain

import (
	"fmt"
	"time"
)

// Task represents a migration task
type Task struct {
	ID          string        `json:"id" bson:"_id"`
	JobID       string        `json:"job_id" bson:"job_id"`
	LastUpdated time.Time     `json:"last_updated" bson:"last_updated"`
	Source      *TaskMetadata `json:"source" bson:"source"`
	State       JobState      `json:"state" bson:"state"`
	Target      *TaskMetadata `json:"target" bson:"target"`
	Type        TaskType      `json:"type" bson:"type"`
	Links       TaskLinks     `json:"links" bson:"links"`
}

// TaskMetadata represents metadata about a task's source or target
type TaskMetadata struct {
	ID    string `json:"id" bson:"id"`
	Label string `json:"label" bson:"label"`
	URI   string `json:"uri" bson:"uri"`
}

// TaskType represents the type of migration task
type TaskType string

const (
	// TaskTypeDataset indicates a dataset task
	TaskTypeDataset TaskType = "dataset"
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
func NewTaskLinks(id, jobID string) TaskLinks {
	return TaskLinks{
		Self: &LinkObject{
			HRef: fmt.Sprintf("/v1/migration-jobs/%s/tasks/%s", jobID, id),
		},
		Job: &LinkObject{
			HRef: fmt.Sprintf("/v1/migration-jobs/%s", jobID),
		},
	}
}
