package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

// Client is a wrapper around the go-slack client that provides
// structured notification sending capabilities
type Client struct {
	client   *slack.Client
	channels Channels
}

// New returns a new Client if Slack notifications are enabled.
// If not enabled, it returns a NoopClient that performs no operations.
// The config is validated before creating the client.
func New(cfg *Config) (Clienter, error) {
	if cfg == nil {
		return nil, errNilSlackConfig
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid slack configuration: %w", err)
	}

	// Return noop client if disabled
	if !cfg.Enabled {
		return &NoopClient{}, nil
	}

	return &Client{
		client:   slack.New(cfg.APIToken),
		channels: cfg.Channels,
	}, nil
}

// SendAlarm sends an error notification to the configured Slack alarm channel.
// The error will be included in the message fields if provided.
func (c *Client) SendAlarm(ctx context.Context, summary string, err error, details map[string]interface{}) error {
	if err := c.validateInput(ctx, summary); err != nil {
		return err
	}

	return c.doSendMessage(
		ctx,
		c.channels.AlarmChannel,
		RedColour,
		AlarmEmoji,
		summary,
		buildAttachmentFields(err, details),
	)
}

// SendWarning sends a warning notification to the configured
// Slack warning channel.
func (c *Client) SendWarning(ctx context.Context, summary string, details map[string]interface{}) error {
	if err := c.validateInput(ctx, summary); err != nil {
		return err
	}

	return c.doSendMessage(
		ctx,
		c.channels.WarningChannel,
		YellowColour,
		WarningEmoji,
		summary,
		buildAttachmentFields(nil, details),
	)
}

// SendInfo sends an info notification to the configured Slack info channel.
func (c *Client) SendInfo(ctx context.Context, summary string, details map[string]interface{}) error {
	if err := c.validateInput(ctx, summary); err != nil {
		return err
	}

	return c.doSendMessage(
		ctx,
		c.channels.InfoChannel,
		GreenColour,
		InfoEmoji,
		summary,
		buildAttachmentFields(nil, details),
	)
}

// validateInput checks that required input parameters are valid
func (c *Client) validateInput(ctx context.Context, summary string) error {
	if ctx == nil {
		return errNilContext
	}
	if summary == "" {
		return errEmptySummary
	}
	return nil
}

// doSendMessage is a helper function to send a message to a specified
// Slack channel with given parameters
func (c *Client) doSendMessage(
	ctx context.Context,
	channel string,
	color Colour,
	emoji Emoji,
	summary string,
	fields []slack.AttachmentField,
) error {
	attachment := slack.Attachment{
		Color:  color.String(),
		Fields: fields,
	}

	_, _, err := c.client.PostMessageContext(
		ctx,
		channel,
		slack.MsgOptionText(
			fmt.Sprintf("%s %s", emoji.String(), summary),
			false,
		),
		slack.MsgOptionAttachments(attachment),
	)

	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}

	return nil
}

// buildAttachmentFields constructs Slack attachment fields from the
// given error and details. If err is not nil, it sets the first field
// to the error message
func buildAttachmentFields(err error, details map[string]interface{}) []slack.AttachmentField {
	fields := []slack.AttachmentField{}

	if err != nil {
		fields = append(fields, slack.AttachmentField{
			Title: "Error",
			Value: err.Error(),
			Short: false,
		})
	}

	for key, value := range details {
		fields = append(fields, slack.AttachmentField{
			Title: key,
			Value: fmt.Sprintf("%v", value),
			Short: true,
		})
	}

	return fields
}
