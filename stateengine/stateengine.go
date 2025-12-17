package stateengine

import (
	"fmt"

	"github.com/ONSdigital/dis-migration-service/domain"
)

// TransitionError is returned when an invalid transition is requested.
type TransitionError struct {
	From domain.State
	To   domain.State
}

// Error returns the error message for a TransitionError.
func (e *TransitionError) Error() string {
	return fmt.Sprintf("invalid state transition: %q -> %q", e.From, e.To)
}

// allowedTransitions encodes the state machine from the design docs.
var allowedTransitions = map[domain.State][]domain.State{
	// happy path
	domain.StateSubmitted:      {domain.StateMigrating, domain.StateCancelled},
	domain.StateMigrating:      {domain.StateInReview, domain.StateFailedMigration},
	domain.StateInReview:       {domain.StateApproved, domain.StateRejected},
	domain.StateApproved:       {domain.StatePublishing},
	domain.StatePublishing:     {domain.StatePublished, domain.StateFailedPublish},
	domain.StatePublished:      {domain.StatePostPublishing},
	domain.StatePostPublishing: {domain.StateCompleted, domain.StateFailedPostPublish},

	// rejection and revert paths
	domain.StateRejected:  {domain.StateReverting},
	domain.StateReverting: {domain.StateCancelled},

	// failure recovery paths
	domain.StateFailedMigration:   {domain.StateRejected},
	domain.StateFailedPublish:     {domain.StateApproved},
	domain.StateFailedPostPublish: {domain.StatePublished},

	// Terminal states have no outgoing transitions
	domain.StateCompleted: {},
	domain.StateCancelled: {},
}

// CanTransition returns true if a transition from `from` to `to` is allowed.
func CanTransition(from, to domain.State) bool {
	nextStates, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	for _, ns := range nextStates {
		if ns == to {
			return true
		}
	}
	return false
}

// ValidateTransition validates a transition or returns a TransitionError.
func ValidateTransition(from, to domain.State) error {
	if !domain.IsValidState(to) {
		return fmt.Errorf("unknown target state %q", to)
	}
	if !domain.IsValidState(from) {
		return fmt.Errorf("unknown current state %q", from)
	}
	if !CanTransition(from, to) {
		return &TransitionError{From: from, To: to}
	}
	return nil
}

// IsAllowedStateForJobUpdate checks if the state is allowed for the
// job update endpoint
func IsAllowedStateForJobUpdate(state domain.State) bool {
	switch state {
	case domain.StateApproved, domain.StateCancelled:
		return true
	default:
		return false
	}
}

// GetNextStates returns all valid next states from the given state.
func GetNextStates(from domain.State) []domain.State {
	if states, ok := allowedTransitions[from]; ok {
		return states
	}
	return []domain.State{}
}

// IsTerminalState returns true if the state has no outgoing transitions.
func IsTerminalState(state domain.State) bool {
	return len(GetNextStates(state)) == 0
}

// IsFailureState returns true if the state represents a failure condition.
func IsFailureState(state domain.State) bool {
	switch state {
	case domain.StateFailedMigration, domain.StateFailedPublish, domain.StateFailedPostPublish:
		return true
	default:
		return false
	}
}

// IsRejectedOrReverting returns true if the state is rejected or reverting.
func IsRejectedOrReverting(state domain.State) bool {
	switch state {
	case domain.StateRejected, domain.StateReverting:
		return true
	default:
		return false
	}
}
