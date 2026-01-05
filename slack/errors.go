package slack

import "errors"

// Configuration validation errors
var (
	errNilSlackConfig        = errors.New("slack configuration is nil")
	errMissingAPIToken       = errors.New("slack API token is missing")
	errMissingInfoChannel    = errors.New("slack info channel is missing")
	errMissingWarningChannel = errors.New("slack warning channel is missing")
	errMissingAlarmChannel   = errors.New("slack alarm channel is missing")
)

// Runtime errors
var (
	errEmptySummary = errors.New("summary cannot be empty")
	errNilContext   = errors.New("context cannot be nil")
)
