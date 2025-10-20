package store

import "errors"

// A list of error messages for the datastore
var (
	ErrJobNotFound = errors.New("job not found")
)
