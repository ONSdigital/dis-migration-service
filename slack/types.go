package slack

// Emoji represents a Slack emoji code
type Emoji string

// String returns the string representation of the emoji
func (e Emoji) String() string {
	return string(e)
}

// Predefined emoji constants for different notification levels
const (
	InfoEmoji    Emoji = ":information_source:"
	WarningEmoji Emoji = ":warning:"
	AlarmEmoji   Emoji = ":rotating_light:"
)

// Colour represents a Slack attachment colour
type Colour string

// String returns the string representation of the colour
func (c Colour) String() string {
	return string(c)
}

// Predefined colour constants matching Slack's attachment colour scheme
const (
	RedColour    Colour = "danger"
	YellowColour Colour = "warning"
	GreenColour  Colour = "good"
)
