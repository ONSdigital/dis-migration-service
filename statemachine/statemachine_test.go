package statemachine_test

import (
	"testing"

	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/statemachine"
)

func TestCanTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     domain.State
		to       domain.State
		expected bool
	}{
		// Happy path transitions
		{
			name:     "submitted to migrating is valid",
			from:     domain.StateSubmitted,
			to:       domain.StateMigrating,
			expected: true,
		},
		{
			name:     "submitted to cancelled is valid",
			from:     domain.StateSubmitted,
			to:       domain.StateCancelled,
			expected: true,
		},
		{
			name:     "migrating to in_review is valid",
			from:     domain.StateMigrating,
			to:       domain.StateInReview,
			expected: true,
		},
		{
			name:     "migrating to failed_migration is valid",
			from:     domain.StateMigrating,
			to:       domain.StateFailedMigration,
			expected: true,
		},
		{
			name:     "in_review to approved is valid",
			from:     domain.StateInReview,
			to:       domain.StateApproved,
			expected: true,
		},
		{
			name:     "in_review to rejected is valid",
			from:     domain.StateInReview,
			to:       domain.StateRejected,
			expected: true,
		},
		{
			name:     "approved to publishing is valid",
			from:     domain.StateApproved,
			to:       domain.StatePublishing,
			expected: true,
		},
		{
			name:     "publishing to published is valid",
			from:     domain.StatePublishing,
			to:       domain.StatePublished,
			expected: true,
		},
		{
			name:     "publishing to failed_publish is valid",
			from:     domain.StatePublishing,
			to:       domain.StateFailedPublish,
			expected: true,
		},
		{
			name:     "published to post_publishing is valid",
			from:     domain.StatePublished,
			to:       domain.StatePostPublishing,
			expected: true,
		},
		{
			name:     "post_publishing to completed is valid",
			from:     domain.StatePostPublishing,
			to:       domain.StateCompleted,
			expected: true,
		},
		{
			name:     "post_publishing to failed_post_publish is valid",
			from:     domain.StatePostPublishing,
			to:       domain.StateFailedPostPublish,
			expected: true,
		},

		// Rejection and revert paths
		{
			name:     "rejected to reverting is valid",
			from:     domain.StateRejected,
			to:       domain.StateReverting,
			expected: true,
		},
		{
			name:     "reverting to cancelled is valid",
			from:     domain.StateReverting,
			to:       domain.StateCancelled,
			expected: true,
		},

		// Failure recovery paths
		{
			name:     "failed_migration to rejected is valid",
			from:     domain.StateFailedMigration,
			to:       domain.StateRejected,
			expected: true,
		},
		{
			name:     "failed_publish to approved is valid (retry endpoint)",
			from:     domain.StateFailedPublish,
			to:       domain.StateApproved,
			expected: true,
		},
		{
			name:     "failed_post_publish to published is valid (retry endpoint)",
			from:     domain.StateFailedPostPublish,
			to:       domain.StatePublished,
			expected: true,
		},

		// Invalid transitions
		{
			name:     "submitted to completed is invalid",
			from:     domain.StateSubmitted,
			to:       domain.StateCompleted,
			expected: false,
		},
		{
			name:     "in_review to reverting is invalid (should go to rejected first)",
			from:     domain.StateInReview,
			to:       domain.StateReverting,
			expected: false,
		},
		{
			name:     "approved to rejected is invalid (should come from in_review)",
			from:     domain.StateApproved,
			to:       domain.StateRejected,
			expected: false,
		},
		{
			name:     "completed to migrating is invalid (terminal state)",
			from:     domain.StateCompleted,
			to:       domain.StateMigrating,
			expected: false,
		},
		{
			name:     "cancelled to submitted is invalid (terminal state)",
			from:     domain.StateCancelled,
			to:       domain.StateSubmitted,
			expected: false,
		},
		{
			name:     "published to publishing is invalid (backward transition)",
			from:     domain.StatePublished,
			to:       domain.StatePublishing,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := statemachine.CanTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf(
					"CanTransition(%q, %q) = %v, want %v",
					tt.from,
					tt.to,
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestValidateTransition(t *testing.T) {
	tests := []struct {
		name        string
		from        domain.State
		to          domain.State
		expectError bool
		errorType   string
	}{
		{
			name:        "valid happy path transition",
			from:        domain.StateSubmitted,
			to:          domain.StateMigrating,
			expectError: false,
		},
		{
			name:        "valid rejection transition",
			from:        domain.StateInReview,
			to:          domain.StateRejected,
			expectError: true,
			errorType:   "TransitionError",
		},
		{
			name:        "invalid transition from terminal state",
			from:        domain.StateCompleted,
			to:          domain.StateMigrating,
			expectError: true,
			errorType:   "TransitionError",
		},
		{
			name:        "unknown target state",
			from:        domain.StateSubmitted,
			to:          domain.State("unknown"),
			expectError: true,
			errorType:   "unknown state",
		},
		{
			name:        "valid retry transition",
			from:        domain.StateFailedPublish,
			to:          domain.StateApproved,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := statemachine.ValidateTransition(tt.from, tt.to)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsTerminalState(t *testing.T) {
	tests := []struct {
		name     string
		state    domain.State
		expected bool
	}{
		{
			name:     "completed is terminal",
			state:    domain.StateCompleted,
			expected: true,
		},
		{
			name:     "cancelled is terminal",
			state:    domain.StateCancelled,
			expected: true,
		},
		{
			name:     "submitted is not terminal",
			state:    domain.StateSubmitted,
			expected: false,
		},
		{
			name:     "migrating is not terminal",
			state:    domain.StateMigrating,
			expected: false,
		},
		{
			name:     "in_review is not terminal",
			state:    domain.StateInReview,
			expected: false,
		},
		{
			name:     "approved is not terminal",
			state:    domain.StateApproved,
			expected: false,
		},
		{
			name:     "publishing is not terminal",
			state:    domain.StatePublishing,
			expected: false,
		},
		{
			name:     "published is not terminal",
			state:    domain.StatePublished,
			expected: false,
		},
		{
			name:     "post_publishing is not terminal",
			state:    domain.StatePostPublishing,
			expected: false,
		},
		{
			name:     "reverting is not terminal",
			state:    domain.StateReverting,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := statemachine.IsTerminalState(tt.state)
			if result != tt.expected {
				t.Errorf(
					"IsTerminalState(%q) = %v, want %v",
					tt.state,
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestGetNextStates(t *testing.T) {
	tests := []struct {
		name     string
		state    domain.State
		expected []domain.State
	}{
		{
			name:     "submitted has two next states",
			state:    domain.StateSubmitted,
			expected: []domain.State{domain.StateMigrating, domain.StateCancelled},
		},
		{
			name:     "migrating has two next states",
			state:    domain.StateMigrating,
			expected: []domain.State{domain.StateInReview, domain.StateFailedMigration},
		},
		{
			name:     "in_review has two next states",
			state:    domain.StateInReview,
			expected: []domain.State{domain.StateApproved, domain.StateRejected},
		},
		{
			name:     "approved has one next state",
			state:    domain.StateApproved,
			expected: []domain.State{domain.StatePublishing},
		},
		{
			name:     "completed has no next states",
			state:    domain.StateCompleted,
			expected: []domain.State{},
		},
		{
			name:     "cancelled has no next states",
			state:    domain.StateCancelled,
			expected: []domain.State{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := statemachine.GetNextStates(tt.state)
			if !statesEqual(result, tt.expected) {
				t.Errorf(
					"GetNextStates(%q) = %v, want %v",
					tt.state,
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestIsFailureState(t *testing.T) {
	tests := []struct {
		name     string
		state    domain.State
		expected bool
	}{
		{
			name:     "failed_migration is failure state",
			state:    domain.StateFailedMigration,
			expected: true,
		},
		{
			name:     "failed_publish is failure state",
			state:    domain.StateFailedPublish,
			expected: true,
		},
		{
			name:     "failed_post_publish is failure state",
			state:    domain.StateFailedPostPublish,
			expected: true,
		},
		{
			name:     "migrating is not failure state",
			state:    domain.StateMigrating,
			expected: false,
		},
		{
			name:     "completed is not failure state",
			state:    domain.StateCompleted,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := statemachine.IsFailureState(tt.state)
			if result != tt.expected {
				t.Errorf(
					"IsFailureState(%q) = %v, want %v",
					tt.state,
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestIsRejectedOrReverting(t *testing.T) {
	tests := []struct {
		name     string
		state    domain.State
		expected bool
	}{
		{
			name:     "rejected is rejected or reverting",
			state:    domain.StateRejected,
			expected: true,
		},
		{
			name:     "reverting is rejected or reverting",
			state:    domain.StateReverting,
			expected: true,
		},
		{
			name:     "approved is not rejected or reverting",
			state:    domain.StateApproved,
			expected: false,
		},
		{
			name:     "cancelled is not rejected or reverting",
			state:    domain.StateCancelled,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := statemachine.IsRejectedOrReverting(tt.state)
			if result != tt.expected {
				t.Errorf(
					"IsRejectedOrReverting(%q) = %v, want %v",
					tt.state,
					result,
					tt.expected,
				)
			}
		})
	}
}

// Helper function to compare state slices
func statesEqual(a, b []domain.State) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
