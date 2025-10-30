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
	ErrInternalServerError = errors.New("an unexpected error occurred")

	StatusCodeMap = map[error]int{
		ErrJobNotFound:         http.StatusNotFound,
		ErrUnableToParseBody:   http.StatusBadRequest,
		ErrSourceIDNotProvided: http.StatusBadRequest,
		ErrTargetIDNotProvided: http.StatusBadRequest,
		ErrJobTypeNotProvided:  http.StatusBadRequest,
		ErrInternalServerError: http.StatusInternalServerError,
	}
)
