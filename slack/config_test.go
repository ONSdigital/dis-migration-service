package slack

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "slack configuration is nil",
		},
		{
			name: "disabled config - no validation needed",
			config: &Config{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "enabled with all fields set",
			config: &Config{
				Enabled:  true,
				APIToken: "xoxb-test-token",
				Channels: Channels{
					InfoChannel:    "#info",
					WarningChannel: "#warning",
					AlarmChannel:   "#alarm",
				},
			},
			wantErr: false,
		},
		{
			name: "enabled with missing API token",
			config: &Config{
				Enabled: true,
				Channels: Channels{
					InfoChannel:    "#info",
					WarningChannel: "#warning",
					AlarmChannel:   "#alarm",
				},
			},
			wantErr: true,
			errMsg:  "API token",
		},
		{
			name: "enabled with missing channels",
			config: &Config{
				Enabled:  true,
				APIToken: "xoxb-test-token",
				Channels: Channels{},
			},
			wantErr: true,
			errMsg:  "info channel",
		},
		{
			name: "enabled with missing alarm channel only",
			config: &Config{
				Enabled:  true,
				APIToken: "xoxb-test-token",
				Channels: Channels{
					InfoChannel:    "#info",
					WarningChannel: "#warning",
				},
			},
			wantErr: true,
			errMsg:  "alarm channel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("Config.Validate() expected error containing %q, got nil", tt.errMsg)
				} else if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Config.Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" ||
		(s != "" && substr != "" && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
