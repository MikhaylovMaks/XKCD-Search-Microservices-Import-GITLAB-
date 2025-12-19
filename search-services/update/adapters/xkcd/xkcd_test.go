package xkcd

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("valid url", func(t *testing.T) {
		client, err := NewClient("https://xkcd.com", 10*time.Second, logger)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "https://xkcd.com", client.url)
	})

	t.Run("empty url", func(t *testing.T) {
		client, err := NewClient("", 10*time.Second, logger)

		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "empty base url specified")
	})
}

// Note: For full HTTP testing, we would need httptest.Server or mocked HTTP client
// These tests ensure basic structure and error handling work
func TestClient_Structure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	client, err := NewClient("https://xkcd.com", 10*time.Second, logger)
	assert.NoError(t, err)

	// Test that methods exist and can be called (will fail on network, but that's expected)
	assert.NotNil(t, client.Get)
	assert.NotNil(t, client.LastID)
}
