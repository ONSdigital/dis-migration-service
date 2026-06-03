package domain

// Action represents the action that was performed given the request to the API
type Action string

// Outcome represents the outcome of the action given the request to the API
type Outcome string

// AuditEventParams represents additional values to include in audit events.
type AuditEventParams map[string]interface{}

const (
	// ActionCreate represents a create action.
	ActionCreate Action = "CREATE"
	// ActionRead represents a read action.
	ActionRead Action = "READ"
	// ActionUpdate represents an update action.
	ActionUpdate Action = "UPDATE"
	// ActionDelete represents a delete action.
	ActionDelete Action = "DELETE"
	// OutcomeSuccess represents a successful outcome.
	OutcomeSuccess Outcome = "success"
	// OutcomeFailure represents a failed outcome.
	OutcomeFailure Outcome = "failure"
)
