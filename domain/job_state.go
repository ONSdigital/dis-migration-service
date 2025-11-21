package domain

// JobState represents the various states a migration job can be in.
type JobState string

const (
	// JobStateSubmitted indicates a job or task has been submitted
	JobStateSubmitted JobState = "submitted"
	// JobStateInReview indicates a job or task is in review
	JobStateInReview JobState = "in_review"
	// JobStateApproved indicates a job or task has been approved
	JobStateApproved JobState = "approved"
	// JobStatePublished indicates a job or task has been published
	JobStatePublished JobState = "published"
	// JobStateCompleted indicates a job or task has completed successfully
	JobStateCompleted JobState = "completed"

	// JobStateMigrating indicates a job or task is currently migrating
	JobStateMigrating JobState = "migrating"
	// JobStatePublishing indicates a job or task is currently publishing
	JobStatePublishing JobState = "publishing"
	// JobStatePostPublishing indicates a job or task is in post-publishing phase
	JobStatePostPublishing JobState = "post_publishing"
	// JobStateReverting indicates a job or task is being reverted
	JobStateReverting JobState = "reverting"

	// JobStateFailedPostPublish indicates a job or task failed
	// during post-publishing
	JobStateFailedPostPublish JobState = "failed_post_publish"
	// JobStateFailedPublish indicates a job or task failed during publishing
	JobStateFailedPublish JobState = "failed_publish"
	// JobStateFailedMigration indicates a job or task failed during migration
	JobStateFailedMigration JobState = "failed_migration"
	// JobStateCancelled indicates a job or task has been cancelled
	JobStateCancelled JobState = "cancelled"
)

// IsValidJobState checks if the provided state is a valid JobState.
func IsValidJobState(state JobState) bool {
	switch state {
	case JobStateSubmitted, JobStateInReview, JobStateApproved, JobStatePublished,
		JobStateCompleted, JobStateMigrating, JobStatePublishing, JobStatePostPublishing,
		JobStateReverting, JobStateFailedMigration, JobStateFailedPostPublish, JobStateFailedPublish,
		JobStateCancelled:
		return true
	default:
		return false
	}
}

// IsValidJobStateForTaskCreation checks if the provided state is a valid
// JobState for task creation.
func IsValidJobStateForTaskCreation(state JobState) bool {
	switch state {
	case JobStateSubmitted, JobStateApproved:
		return true
	default:
		return false
	}
}

// GetNonCancelledStates returns a slice of JobStates excluding the
// 'cancelled' state.
func GetNonCancelledStates() []JobState {
	return []JobState{
		JobStateSubmitted, JobStateInReview, JobStateApproved, JobStatePublished,
		JobStateCompleted, JobStateMigrating, JobStatePublishing, JobStatePostPublishing,
		JobStateReverting, JobStateFailedMigration, JobStateFailedPostPublish, JobStateFailedPublish,
	}
}
