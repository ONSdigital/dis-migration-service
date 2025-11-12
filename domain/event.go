package domain

// Event represents an action taken on a migration job
type Event struct {
	CreatedAt   string `json:"created_at"`
	RequestedBy User   `json:"requested_by"`
	Action      string `json:"action"`
	JobID       string `json:"job_id"`
}

// User represents a user who has performed an action
// on a migration job
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}
