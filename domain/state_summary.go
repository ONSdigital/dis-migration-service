package domain

// StateSummary represents a summary count for a given state.
type StateSummary struct {
	ID    State  `json:"id" bson:"_id"`
	Label string `json:"label" bson:"label"`
	Count int    `json:"count" bson:"count"`
}
