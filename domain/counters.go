package domain

// Counter represents a counter relating to migration
// e.g. a counter for creating new job numbers
type Counter struct {
	CounterName  string `json:"counter_name" bson:"counter_name"`
	CounterValue int    `json:"counter_value" bson:"counter_value"`
}
