package errors

import (
	"errors"
	"net/http"
)

// ErrorList represents a list of errors.
type ErrorList struct {
	Errors []Error `json:"errors"`
}

// Error represents a single error with a status code and description.
type Error struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

// New creates a new redacted Error based on the provided error.
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
	ErrJobNumberCounterNotFound  = errors.New("job number counter not found")
	ErrJobNotFound               = errors.New("job not found")
	ErrUnableToParseBody         = errors.New("unable to read submitted body")
	ErrJobNumberNotProvided      = errors.New("job number not provided")
	ErrJobNumberMustBeInt        = errors.New("job number must be an integer")
	ErrSourceIDNotProvided       = errors.New("source ID not provided")
	ErrTargetIDNotProvided       = errors.New("target ID not provided")
	ErrJobTypeNotProvided        = errors.New("job type not provided")
	ErrSourceTitleNotFound       = errors.New("source title not found or empty")
	ErrJobIDNotProvided             = errors.New("job ID not provided")
	ErrSourceIDInvalid              = errors.New("source ID is invalid")
	ErrTargetIDInvalid              = errors.New("target ID is invalid")
	ErrJobTypeInvalid               = errors.New("job type is invalid")
	ErrInternalServerError          = errors.New("an unexpected error occurred")
	ErrSourceIDValidation           = errors.New("source ID failed to validate")
	ErrTargetIDValidation           = errors.New("target ID failed to validate")
	ErrJobAlreadyRunning            = errors.New("job already running")
	ErrJobStateInvalid              = errors.New("job state parameter is invalid")
	ErrTaskStateInvalid             = errors.New("task state parameter is invalid")
	ErrJobStateNotAllowed           = errors.New("state not allowed for this endpoint")
	ErrJobStateTransitionNotAllowed = errors.New("state change is not allowed")
	ErrTaskNotFound                 = errors.New("task not found")
	ErrOffsetInvalid                = errors.New("offset parameter is invalid")
	ErrLimitInvalid                 = errors.New("limit parameter is invalid")
	ErrLimitExceeded                = errors.New("limit parameter exceeds maximum allowed")

	ErrSourceIDZebedeeURIInvalid = errors.New("source ID URI path must start with '/', not end with '/', not contain query strings or hashbangs")
	ErrTargetIDDatasetIDInvalid  = errors.New("target id must be lowercase alphanumeric with optional hyphen separators")

	StatusCodeMap = map[error]int{
		ErrJobNotFound:               http.StatusNotFound,
		ErrUnableToParseBody:         http.StatusBadRequest,
		ErrJobNumberNotProvided:      http.StatusBadRequest,
		ErrSourceIDNotProvided:       http.StatusBadRequest,
		ErrTargetIDNotProvided:       http.StatusBadRequest,
		ErrJobTypeNotProvided:        http.StatusBadRequest,
		ErrInternalServerError:       http.StatusInternalServerError,
		ErrSourceTitleNotFound:       http.StatusInternalServerError,
		ErrSourceIDValidation:        http.StatusInternalServerError,
		ErrTargetIDValidation:        http.StatusInternalServerError,
		ErrJobAlreadyRunning:         http.StatusConflict,
		ErrSourceIDInvalid:           http.StatusBadRequest,
		ErrTargetIDInvalid:           http.StatusBadRequest,
		ErrJobTypeInvalid:            http.StatusBadRequest,
		ErrSourceIDZebedeeURIInvalid: http.StatusBadRequest,
		ErrTargetIDDatasetIDInvalid:  http.StatusBadRequest,
		ErrJobStateInvalid:           http.StatusBadRequest,
		ErrTaskStateInvalid:          http.StatusBadRequest,
		ErrOffsetInvalid:             http.StatusBadRequest,
		ErrLimitInvalid:              http.StatusBadRequest,
		ErrLimitExceeded:             http.StatusBadRequest,
		ErrJobStateTransitionNotAllowed: http.StatusConflict,
		ErrSourceIDInvalid:              http.StatusBadRequest,
		ErrTargetIDInvalid:              http.StatusBadRequest,
		ErrJobTypeInvalid:               http.StatusBadRequest,
		ErrJobStateNotAllowed:           http.StatusBadRequest,
	}
)
