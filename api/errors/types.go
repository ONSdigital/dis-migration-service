package errors

import "net/http"

// A list of error messages for Migration Service
var (
	ErrJobNotFound         = NewAPIError(http.StatusNotFound, "job not found")
	ErrInternalServerError = NewAPIError(http.StatusInternalServerError, "internal server error")
	ErrUnableToParseBody   = NewAPIError(http.StatusBadRequest, "unable to read submitted body")
	ErrSourceIDNotProvided = NewAPIError(http.StatusBadRequest, "source ID not provided")
	ErrTargetIDNotProvided = NewAPIError(http.StatusBadRequest, "target ID not provided")
	ErrJobTypeNotProvided  = NewAPIError(http.StatusBadRequest, "job type not provided")
)
