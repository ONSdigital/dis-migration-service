package domain

import (
	"fmt"
)

// State represents the various states a migration job can be in.
type State string

const (
	// StateSubmitted indicates a job or task has been submitted
	StateSubmitted State = "submitted"
	// StateInReview indicates a job or task is in review
	StateInReview State = "in_review"
	// StateApproved indicates a job or task has been approved
	StateApproved State = "approved"
	// StateRejected indicates a job or task has been rejected
	StateRejected State = "rejected"
	// StatePublished indicates a job or task has been published
	StatePublished State = "published"
	// StateCompleted indicates a job or task has completed successfully
	StateCompleted State = "completed"

	// StateMigrating indicates a job or task is currently migrating
	StateMigrating State = "migrating"
	// StatePublishing indicates a job or task is currently publishing
	StatePublishing State = "publishing"
	// StatePendingPostPublish indicates a task is pending post-publishing steps
	StatePendingPostPublish State = "pending_post_publish"
	// StatePostPublishing indicates a job or task is in post-publishing phase
	StatePostPublishing State = "post_publishing"
	// StateReverting indicates a job or task is being reverted
	StateReverting State = "reverting"

	// StateFailedPostPublish indicates a job or task failed
	// during post-publishing
	StateFailedPostPublish State = "failed_post_publish"
	// StateFailedPublish indicates a job or task failed during publishing
	StateFailedPublish State = "failed_publish"
	// StateFailedMigration indicates a job or task failed during migration
	StateFailedMigration State = "failed_migration"
	// StateFailedReversion indicates a job or task failed during reverting
	StateFailedReversion State = "failed_reversion"
	// StateCancelled indicates a job or task has been cancelled
	StateCancelled State = "cancelled"
)

// IsValidState checks if the provided state is a valid State.
func IsValidState(state State) bool {
	switch state {
	case StateSubmitted, StateInReview, StateApproved, StateRejected, StatePublished,
		StateCompleted, StateMigrating, StatePublishing, StatePendingPostPublish, StatePostPublishing,
		StateReverting, StateFailedMigration, StateFailedPostPublish, StateFailedPublish, StateFailedReversion,
		StateCancelled:
		return true
	default:
		return false
	}
}

// GetNonCancelledStates returns a slice of States excluding the
// 'cancelled' state.
func GetNonCancelledStates() []State {
	return []State{
		StateSubmitted, StateInReview, StateApproved, StateRejected, StatePublished,
		StateCompleted, StateMigrating, StatePublishing, StatePendingPostPublish, StatePostPublishing,
		StateReverting, StateFailedMigration, StateFailedPostPublish, StateFailedPublish, StateFailedReversion,
	}
}

// IsFailedState checks if the provided state is a failure state.
func IsFailedState(state State) bool {
	switch state {
	case StateFailedMigration, StateFailedPostPublish, StateFailedPublish, StateFailedReversion:
		return true
	default:
		return false
	}
}

// GetStateLabel returns a human readable label for the given State.
func GetStateLabel(state State) (string, error) {
	switch state {
	case StateSubmitted:
		return "Submitted", nil
	case StateInReview:
		return "In review", nil
	case StateApproved:
		return "Approved", nil
	case StateRejected:
		return "Rejected", nil
	case StatePublished:
		return "Published", nil
	case StateCompleted:
		return "Completed", nil
	case StateMigrating:
		return "Migrating", nil
	case StatePublishing:
		return "Publishing", nil
	case StatePendingPostPublish:
		return "Pending post-publish", nil
	case StatePostPublishing:
		return "Post-publishing", nil
	case StateReverting:
		return "Reverting", nil
	case StateFailedPostPublish:
		return "Failed post-publish", nil
	case StateFailedPublish:
		return "Failed publish", nil
	case StateFailedMigration:
		return "Failed migration", nil
	case StateFailedReversion:
		return "Failed reversion", nil
	case StateCancelled:
		return "Cancelled", nil
	default:
		return "", fmt.Errorf("unknown state label mapping for state: %s", state)
	}
}
