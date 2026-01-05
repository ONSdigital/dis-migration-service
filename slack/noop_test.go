package slack

import (
	"context"
	"errors"
	"testing"
)

func TestNoopClient_SendAlarm(t *testing.T) {
	client := &NoopClient{}
	err := client.SendAlarm(
		context.Background(),
		"test",
		errors.New("test error"),
		map[string]interface{}{"key": "value"},
	)
	if err != nil {
		t.Errorf("NoopClient.SendAlarm() error = %v, want nil", err)
	}
}

func TestNoopClient_SendWarning(t *testing.T) {
	client := &NoopClient{}
	err := client.SendWarning(
		context.Background(),
		"test",
		map[string]interface{}{"key": "value"},
	)
	if err != nil {
		t.Errorf("NoopClient.SendWarning() error = %v, want nil", err)
	}
}

func TestNoopClient_SendInfo(t *testing.T) {
	client := &NoopClient{}
	err := client.SendInfo(
		context.Background(),
		"test",
		map[string]interface{}{"key": "value"},
	)
	if err != nil {
		t.Errorf("NoopClient.SendInfo() error = %v, want nil", err)
	}
}
