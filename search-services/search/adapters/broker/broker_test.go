package broker

import (
	"testing"

	"log/slog"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, nil))

	t.Run("connection failure with invalid address", func(t *testing.T) {
		client, err := NewClient("nats://invalid:1234", logger)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestClient_Close(t *testing.T) {
	t.Run("close with nil connection", func(t *testing.T) {
		client := &Client{}
		err := client.Close()
		assert.NoError(t, err)
	})

	t.Run("close with logger", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(nil, nil))
		client := &Client{log: logger}
		err := client.Close()
		assert.NoError(t, err)
	})
}

// Note: NewClient with valid connection and Subscribe require a running NATS server.
// For unit tests, we test the failure cases and basic functionality.
// Integration tests would cover the full functionality with a test server.
