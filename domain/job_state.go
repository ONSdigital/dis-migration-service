package domain

type JobState string

const (
	JobStateSubmitted         JobState = "submitted"
	JobStateInReview          JobState = "in_review"
	JobStateApproved          JobState = "approved"
	JobStatePublished         JobState = "published"
	JobStateCompleted         JobState = "completed"
	JobStateMigrating         JobState = "migrating"
	JobStatePublishing        JobState = "publishing"
	JobStatePostPublishing    JobState = "post_publishing"
	JobStateReverting         JobState = "reverting"
	JobStateFailedPostPublish JobState = "failed_post_publish"
	JobStateFailedPublish     JobState = "failed_publish"
	JobStateFailedMigration   JobState = "failed_migration"
	JobStateCancelled         JobState = "cancelled"
)

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

func GetNonCancelledStates() []JobState {
	return []JobState{
		JobStateSubmitted, JobStateInReview, JobStateApproved, JobStatePublished,
		JobStateCompleted, JobStateMigrating, JobStatePublishing, JobStatePostPublishing,
		JobStateReverting, JobStateFailedMigration, JobStateFailedPostPublish, JobStateFailedPublish,
	}
}
