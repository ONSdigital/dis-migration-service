package domain

// LinkObject represents a generic structure for all links
type LinkObject struct {
	HRef      string `bson:"href,omitempty"  json:"href,omitempty"`
	JobNumber int    `bson:"job_number,omitempty"    json:"job_number,omitempty"`
}
