package domain

import "fmt"

// TaskState represents the various states a migration job can be in.
type TaskState string

const (
	// TaskStateSubmitted indicates a job or task has been submitted
	TaskStateSubmitted TaskState = "submitted"
	// TaskStateInReview indicates a job or task is in review
	TaskStateInReview TaskState = "in_review"
	// TaskStateApproved indicates a job or task has been approved
	TaskStateApproved TaskState = "approved"
	// TaskStatePublished indicates a job or task has been published
	TaskStatePublished TaskState = "published"
	// TaskStateCompleted indicates a job or task has completed successfully
	TaskStateCompleted TaskState = "completed"

	// TaskStateMigrating indicates a job or task is currently migrating
	TaskStateMigrating TaskState = "migrating"
	// TaskStatePublishing indicates a job or task is currently publishing
	TaskStatePublishing TaskState = "publishing"
	// TaskStatePostPublishing indicates a job or task is in post-publishing phase
	TaskStatePostPublishing TaskState = "post_publishing"
	// TaskStateReverting indicates a job or task is being reverted
	TaskStateReverting TaskState = "reverting"

	// TaskStateFailedPostPublish indicates a job or task failed
	// during post-publishing
	TaskStateFailedPostPublish TaskState = "failed_post_publish"
	// TaskStateFailedPublish indicates a job or task failed during publishing
	TaskStateFailedPublish TaskState = "failed_publish"
	// TaskStateFailedMigration indicates a job or task failed during migration
	TaskStateFailedMigration TaskState = "failed_migration"
	// TaskStateCancelled indicates a job or task has been cancelled
	TaskStateCancelled TaskState = "cancelled"
)

var (
	// taskFailureStateMap maps active task states to their
	// corresponding failure states
	taskFailureStateMap = map[TaskState]TaskState{
		TaskStateMigrating:      TaskStateFailedMigration,
		TaskStatePublishing:     TaskStateFailedPublish,
		TaskStatePostPublishing: TaskStateFailedPostPublish,
	}
)

// IsValidTaskState checks if the provided state is a valid TaskState.
func IsValidTaskState(state TaskState) bool {
	switch state {
	case TaskStateSubmitted, TaskStateInReview, TaskStateApproved, TaskStatePublished,
		TaskStateCompleted, TaskStateMigrating, TaskStatePublishing, TaskStatePostPublishing,
		TaskStateReverting, TaskStateFailedMigration, TaskStateFailedPostPublish, TaskStateFailedPublish,
		TaskStateCancelled:
		return true
	default:
		return false
	}
}

// GetFailureStateForTaskState returns the corresponding failure state
// for a given active task state.
func GetFailureStateForTaskState(state TaskState) (TaskState, error) {
	failureState, exists := taskFailureStateMap[state]
	if !exists {
		return "", fmt.Errorf("no failure state defined for task state: %s", state)
	}
	return failureState, nil
}
