package domain

import "fmt"

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
	// StateCancelled indicates a job or task has been cancelled
	StateCancelled State = "cancelled"
)

// IsValidState checks if the provided state is a valid State.
func IsValidState(state State) bool {
	switch state {
	case StateSubmitted, StateInReview, StateApproved, StateRejected, StatePublished,
		StateCompleted, StateMigrating, StatePublishing, StatePostPublishing,
		StateReverting, StateFailedMigration, StateFailedPostPublish, StateFailedPublish,
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
		StateSubmitted, StateInReview, StateApproved, StatePublished,
		StateCompleted, StateMigrating, StatePublishing, StatePostPublishing,
		StateReverting, StateFailedMigration, StateFailedPostPublish, StateFailedPublish,
	}
}

// IsFailedState checks if the provided state is a failure state.
func IsFailedState(state State) bool {
	switch state {
	case StateFailedMigration, StateFailedPostPublish, StateFailedPublish:
		return true
	default:
		return false
	}
}

var (
	// jobFailureStateMap maps active task states to their corresponding
	// failure states
	jobFailureStateMap = map[State]State{
		StateMigrating:      StateFailedMigration,
		StatePublishing:     StateFailedPublish,
		StatePostPublishing: StateFailedPostPublish,
	}
)

// GetFailureStateForJobState returns the corresponding failure state
// for a given active job state.
func GetFailureStateForJobState(state State) (State, error) {
	failureState, exists := jobFailureStateMap[state]
	if !exists {
		return "", fmt.Errorf("no failure state defined for job state: %s", state)
	}
	return failureState, nil
}
