package slack

import (
	"context"
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		wantNoop    bool
		wantErr     bool
		errContains string
	}{
		{
			name:        "nil config",
			config:      nil,
			wantErr:     true,
			errContains: "slack configuration is nil",
		},
		{
			name: "disabled config returns noop",
			config: &Config{
				Enabled: false,
			},
			wantNoop: true,
			wantErr:  false,
		},
		{
			name: "enabled with invalid config",
			config: &Config{
				Enabled: true,
			},
			wantErr:     true,
			errContains: "invalid slack configuration",
		},
		{
			name: "enabled with valid config",
			config: &Config{
				Enabled:  true,
				APIToken: "xoxb-test-token",
				Channels: Channels{
					InfoChannel:    "#info",
					WarningChannel: "#warning",
					AlarmChannel:   "#alarm",
				},
			},
			wantNoop: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("New() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if !tt.wantErr {
				if client == nil {
					t.Error("New() returned nil client when error not expected")
					return
				}

				_, isNoop := client.(*NoopClient)
				if isNoop != tt.wantNoop {
					t.Errorf("New() returned noop = %v, want %v", isNoop, tt.wantNoop)
				}
			}
		})
	}
}

func TestClient_validateInput(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name    string
		ctx     context.Context
		summary string
		wantErr error
	}{
		{
			name:    "valid input",
			ctx:     context.Background(),
			summary: "test summary",
			wantErr: nil,
		},
		{
			name:    "nil context",
			ctx:     nil,
			summary: "test summary",
			wantErr: errNilContext,
		},
		{
			name:    "empty summary",
			ctx:     context.Background(),
			summary: "",
			wantErr: errEmptySummary,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.validateInput(tt.ctx, tt.summary)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("validateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildAttachmentFields(t *testing.T) {
	testErr := errors.New("test error")

	tests := []struct {
		name           string
		err            error
		details        map[string]interface{}
		wantFieldCount int
		wantErrorField bool
	}{
		{
			name:           "no error, no details",
			err:            nil,
			details:        nil,
			wantFieldCount: 0,
			wantErrorField: false,
		},
		{
			name:           "with error, no details",
			err:            testErr,
			details:        nil,
			wantFieldCount: 1,
			wantErrorField: true,
		},
		{
			name: "no error, with details",
			err:  nil,
			details: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
			wantFieldCount: 2,
			wantErrorField: false,
		},
		{
			name: "with error and details",
			err:  testErr,
			details: map[string]interface{}{
				"key1": "value1",
			},
			wantFieldCount: 2,
			wantErrorField: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := buildAttachmentFields(tt.err, tt.details)

			if len(fields) != tt.wantFieldCount {
				t.Errorf("buildAttachmentFields() returned %d fields, want %d", len(fields), tt.wantFieldCount)
			}

			hasErrorField := false
			for _, field := range fields {
				if field.Title == "Error" {
					hasErrorField = true
					if field.Value != testErr.Error() {
						t.Errorf("Error field value = %v, want %v", field.Value, testErr.Error())
					}
				}
			}

			if hasErrorField != tt.wantErrorField {
				t.Errorf("buildAttachmentFields() has error field = %v, want %v", hasErrorField, tt.wantErrorField)
			}
		})
	}
}
