package application

import "errors"

var (
	ErrJobAlreadyRunning  = errors.New("job already running")
	ErrSourceIDNotFound   = errors.New("source id not found")
	ErrSourceIDValidation = errors.New("source id validation error")

	ErrTargetIDFound      = errors.New("target ID found")
	ErrTargetIDValidation = errors.New("target id validation error")
)
