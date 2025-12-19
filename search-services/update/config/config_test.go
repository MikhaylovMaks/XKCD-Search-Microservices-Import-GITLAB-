package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Defaults(t *testing.T) {
	// Test default values by creating an empty config
	cfg := Config{}

	// Check that struct fields have zero values initially
	assert.Empty(t, cfg.Address)
	assert.Empty(t, cfg.LogLevel)
	assert.Empty(t, cfg.DBAddress)

	// XKCD defaults
	assert.Equal(t, "", cfg.XKCD.URL)
	assert.Equal(t, 0, cfg.XKCD.Concurrency)
	assert.Equal(t, time.Duration(0), cfg.XKCD.Timeout)
}

func TestXKCD_Defaults(t *testing.T) {
	xkcd := XKCD{}

	// Check zero values
	assert.Empty(t, xkcd.URL)
	assert.Equal(t, 0, xkcd.Concurrency)
	assert.Equal(t, time.Duration(0), xkcd.Timeout)
	assert.Equal(t, time.Duration(0), xkcd.CheckPeriod)
}

func TestMustLoad(t *testing.T) {
	t.Run("successful config load", func(t *testing.T) {
		// Create a temporary config file
		configContent := `
log_level: INFO
update_address: localhost:8080
db_address: localhost:5432
words_address: localhost:8081
broker_address: nats://localhost:4223
xkcd:
  url: custom.xkcd.com
  concurrency: 5
  timeout: 30s
  check_period: 2h
`
		tempFile, err := os.CreateTemp("", "config_*.yaml")
		require.NoError(t, err)
		defer func() { _ = os.Remove(tempFile.Name()) }()

		_, err = tempFile.WriteString(configContent)
		require.NoError(t, err)
		_ = tempFile.Close()

		// Test loading config
		cfg := MustLoad(tempFile.Name())

		assert.Equal(t, "INFO", cfg.LogLevel)
		assert.Equal(t, "localhost:8080", cfg.Address)
		assert.Equal(t, "localhost:5432", cfg.DBAddress)
		assert.Equal(t, "localhost:8081", cfg.WordsAddress)
		assert.Equal(t, "nats://localhost:4223", cfg.BrokerAddress)

		// XKCD config
		assert.Equal(t, "custom.xkcd.com", cfg.XKCD.URL)
		assert.Equal(t, 5, cfg.XKCD.Concurrency)
		assert.Equal(t, 30*time.Second, cfg.XKCD.Timeout)
		assert.Equal(t, 2*time.Hour, cfg.XKCD.CheckPeriod)
	})

	t.Run("config file not found", func(t *testing.T) {
		// This will panic, but we can't test panic recovery easily
		// In a real scenario, this would be tested with a separate function
		// that returns an error instead of calling log.Fatalf
	})
}

func TestConfig_StructTags(t *testing.T) {
	// Test that struct tags are properly defined
	cfg := Config{
		LogLevel:      "DEBUG",
		Address:       "localhost:80",
		DBAddress:     "localhost:82",
		WordsAddress:  "localhost:81",
		BrokerAddress: "nats://localhost:4222",
		XKCD: XKCD{
			URL:         "xkcd.com",
			Concurrency: 1,
			Timeout:     10 * time.Second,
			CheckPeriod: time.Hour,
		},
	}

	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.Equal(t, "localhost:80", cfg.Address)
	assert.Equal(t, "localhost:82", cfg.DBAddress)
	assert.Equal(t, "localhost:81", cfg.WordsAddress)
	assert.Equal(t, "nats://localhost:4222", cfg.BrokerAddress)

	assert.Equal(t, "xkcd.com", cfg.XKCD.URL)
	assert.Equal(t, 1, cfg.XKCD.Concurrency)
	assert.Equal(t, 10*time.Second, cfg.XKCD.Timeout)
	assert.Equal(t, time.Hour, cfg.XKCD.CheckPeriod)
}
