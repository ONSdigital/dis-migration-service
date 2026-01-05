package slack

// Config holds configuration for sending Slack notifications
type Config struct {
	// Enabled determines whether Slack notifications are active
	Enabled bool `envconfig:"SLACK_ENABLED" default:"false"`

	// APIToken is the Slack bot token for authentication
	APIToken string `envconfig:"SLACK_API_TOKEN"`

	// Channels holds the Slack channel names for different notification levels
	Channels Channels
}

// Channels holds the Slack channel names for different notification levels
type Channels struct {
	InfoChannel    string `envconfig:"SLACK_INFO_CHANNEL"`
	WarningChannel string `envconfig:"SLACK_WARNING_CHANNEL"`
	AlarmChannel   string `envconfig:"SLACK_ALARM_CHANNEL"`
}

// Validate checks that all required fields are set in the Config
// when Slack notifications are enabled
func (c *Config) Validate() error {
	if c == nil {
		return errNilSlackConfig
	}

	// If Slack is not enabled, no validation is required
	if !c.Enabled {
		return nil
	}

	// When enabled, validate all required fields
	if c.APIToken == "" {
		return errMissingAPIToken
	}
	if c.Channels.InfoChannel == "" {
		return errMissingInfoChannel
	}
	if c.Channels.WarningChannel == "" {
		return errMissingWarningChannel
	}
	if c.Channels.AlarmChannel == "" {
		return errMissingAlarmChannel
	}

	return nil
}
