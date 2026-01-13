package slack

import (
	"context"
)

//go:generate moq -out mocks/client.go -pkg mocks . Clienter

// Clienter represents an interface for sending Slack notifications
// across different severity levels.
type Clienter interface {
	// SendAlarm sends a critical error notification to the alarm channel
	SendAlarm(ctx context.Context, summary string, err error, details SlackDetails) error

	// SendWarning sends a warning notification to the warning channel
	SendWarning(ctx context.Context, summary string, details SlackDetails) error

	// SendInfo sends an informational notification to the info channel
	SendInfo(ctx context.Context, summary string, details SlackDetails) error
}
