package domain

type Event struct {
	CreatedAt   string `json:"created_at"`
	RequestedBy User   `json:"requested_by"`
	Action      string `json:"action"`
	JobID       string `json:"job_id"`
}

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}
