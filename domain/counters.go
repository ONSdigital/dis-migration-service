package domain

// Counters represents a set of counter relating to migration
type Counters struct {
	JobNumberCounter string `json:"job_number_counter" bson:"job_number_counter"`
}
