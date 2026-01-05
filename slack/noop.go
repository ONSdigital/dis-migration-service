package slack

import "context"

// NoopClient is a Client that performs no operations.
// It is used when Slack notifications are disabled to avoid
// runtime overhead and maintain clean code paths without nil checks.
type NoopClient struct{}

// SendAlarm is a no-op implementation
func (n *NoopClient) SendAlarm(ctx context.Context, summary string, err error, details map[string]interface{}) error {
	return nil
}

// SendWarning is a no-op implementation
func (n *NoopClient) SendWarning(ctx context.Context, summary string, details map[string]interface{}) error {
	return nil
}

// SendInfo is a no-op implementation
func (n *NoopClient) SendInfo(ctx context.Context, summary string, details map[string]interface{}) error {
	return nil
}
