package errors

import (
	"errors"
	"net/http"
)

type ErrorList struct {
	Errors []Error `json:"errors"`
}

type Error struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

func New(err error) Error {
	var redactedError Error
	code, exists := StatusCodeMap[err]
	if exists {
		redactedError.Code = code
		redactedError.Description = err.Error()
	} else {
		// Mask any unknown errors
		redactedError.Code = http.StatusInternalServerError
		redactedError.Description = ErrInternalServerError.Error()
	}
	return redactedError
}

// Predefined errors
var (
	ErrJobNotFound         = errors.New("job not found")
	ErrUnableToParseBody   = errors.New("unable to read submitted body")
	ErrSourceIDNotProvided = errors.New("source ID not provided")
	ErrTargetIDNotProvided = errors.New("target ID not provided")
	ErrJobTypeNotProvided  = errors.New("job type not provided")
	ErrSourceIDInvalid     = errors.New("source ID is invalid")
	ErrTargetIDInvalid     = errors.New("target ID is invalid")
	ErrJobTypeInvalid      = errors.New("job type is invalid")
	ErrInternalServerError = errors.New("an unexpected error occurred")
	ErrSourceIDValidation  = errors.New("source ID failed to validate")
	ErrTargetIDValidation  = errors.New("target ID failed to validate")
	ErrJobAlreadyRunning   = errors.New("job already running")

	StatusCodeMap = map[error]int{
		ErrJobNotFound:         http.StatusNotFound,
		ErrUnableToParseBody:   http.StatusBadRequest,
		ErrSourceIDNotProvided: http.StatusBadRequest,
		ErrTargetIDNotProvided: http.StatusBadRequest,
		ErrJobTypeNotProvided:  http.StatusBadRequest,
		ErrInternalServerError: http.StatusInternalServerError,
		ErrSourceIDValidation:  http.StatusInternalServerError,
		ErrTargetIDValidation:  http.StatusInternalServerError,
		ErrJobAlreadyRunning:   http.StatusConflict,
		ErrSourceIDInvalid:     http.StatusBadRequest,
		ErrTargetIDInvalid:     http.StatusBadRequest,
		ErrJobTypeInvalid:      http.StatusBadRequest,
	}
)
