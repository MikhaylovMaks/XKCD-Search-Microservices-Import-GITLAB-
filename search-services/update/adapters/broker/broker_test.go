package broker

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("invalid address", func(t *testing.T) {
		client, err := NewClient("invalid://address", logger)

		// Should fail to connect
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestClient_Structure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Test that we can create a Client instance (even if connection fails)
	client := &Client{
		log: logger,
		nc:  nil,
	}

	// Test that methods exist and can be called
	assert.NotNil(t, client.Close)
	assert.NotNil(t, client.Publish)

	// Test Close with nil connection (should not panic)
	err := client.Close()
	assert.NoError(t, err)
}
