package domain

import (
	"fmt"
	"time"
)

// Task represents a migration task
type Task struct {
	ID          string            `json:"id" bson:"_id"`
	JobID       string            `json:"job_id" bson:"job_id"`
	LastUpdated time.Time         `json:"last_updated" bson:"last_updated"`
	Source      *TaskMetadata     `json:"source" bson:"source"`
	State       MigrationState    `json:"state" bson:"state"`
	Target      *TaskMetadata     `json:"target" bson:"target"`
	Type        MigrationTaskType `json:"type" bson:"type"`
	Links       TaskLinks         `json:"links" bson:"links"`
}

// TaskMetadata represents metadata about a task's source or target
type TaskMetadata struct {
	ID    string `json:"id" bson:"id"`
	Label string `json:"label" bson:"label"`
	URI   string `json:"uri" bson:"uri"`
}

// MigrationTaskType represents the type of migration task
type MigrationTaskType string

const (
	// MigrationTaskTypeDataset indicates a dataset task
	MigrationTaskTypeDataset MigrationTaskType = "dataset"
	// MigrationTaskTypeDatasetEdition indicates a dataset edition task
	MigrationTaskTypeDatasetEdition MigrationTaskType = "dataset_edition"
	// MigrationTaskTypeDatasetVersion indicates a dataset version task
	MigrationTaskTypeDatasetVersion MigrationTaskType = "dataset_version"
	// MigrationTaskTypeDatasetDownload indicates a dataset download task
	MigrationTaskTypeDatasetDownload MigrationTaskType = "dataset_download"
)

// MigrationState represents the state of a migration job or task
type MigrationState string

const (
	// MigrationStateSubmitted indicates a job or task has been submitted
	MigrationStateSubmitted MigrationState = "submitted"
	// MigrationStateInReview indicates a job or task is in review
	MigrationStateInReview MigrationState = "in_review"
	// MigrationStateApproved indicates a job or task has been approved
	MigrationStateApproved MigrationState = "approved"
	// MigrationStatePublished indicates a job or task has been published
	MigrationStatePublished MigrationState = "published"
	// MigrationStateCompleted indicates a job or task has completed successfully
	MigrationStateCompleted MigrationState = "completed"

	// MigrationStateMigrating indicates a job or task is currently migrating
	MigrationStateMigrating MigrationState = "migrating"
	// MigrationStatePublishing indicates a job or task is currently publishing
	MigrationStatePublishing MigrationState = "publishing"
	// MigrationStatePostPublishing indicates a job or task is
	// in post-publishing phase
	MigrationStatePostPublishing MigrationState = "post_publishing"
	// MigrationStateReverting indicates a job or task is being reverted
	MigrationStateReverting MigrationState = "reverting"

	// MigrationStateFailedPostPublish indicates a job or task failed
	// during post-publishing
	MigrationStateFailedPostPublish MigrationState = "failed_post_publish"
	// MigrationStateFailedPublish indicates a job or task failed during publishing
	MigrationStateFailedPublish MigrationState = "failed_publish"
	// MigrationStateFailedMigration indicates a job or task failed during migration
	MigrationStateFailedMigration MigrationState = "failed_migration"
	// MigrationStateCancelled indicates a job or task has been cancelled
	MigrationStateCancelled MigrationState = "cancelled"
)

// TaskLinks contains HATEOAS links for a migration task
type TaskLinks struct {
	Self *LinkObject `bson:"self,omitempty" json:"self,omitempty"`
	Job  *LinkObject `bson:"job,omitempty" json:"job,omitempty"`
}

// NewTaskLinks creates TaskLinks for a task
func NewTaskLinks(id, jobID, host string) TaskLinks {
	return TaskLinks{
		Self: &LinkObject{
			HRef: fmt.Sprintf("%s/v1/migration-jobs/%s/tasks/%s", host, jobID, id),
		},
		Job: &LinkObject{
			HRef: fmt.Sprintf("%s/v1/migration-jobs/%s", host, jobID),
		},
	}
}
