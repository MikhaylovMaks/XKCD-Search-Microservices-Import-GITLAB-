package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Defaults(t *testing.T) {
	cfg := Config{}

	assert.Empty(t, cfg.Address)
	assert.Empty(t, cfg.LogLevel)
	assert.Equal(t, time.Duration(0), cfg.IndexTTL)

	assert.Empty(t, cfg.DBAddress)
	assert.Empty(t, cfg.WordsAddress)
	assert.Empty(t, cfg.BrokerAddress)
}

func TestMustLoad(t *testing.T) {
	t.Run("successful config load", func(t *testing.T) {

		configContent := `
log_level: INFO
index_ttl: 2h
search_address: localhost:8080
db_address: localhost:5432
words_address: localhost:8081
broker_address: nats://localhost:4223
`
		tempFile, err := os.CreateTemp("", "config_*.yaml")
		require.NoError(t, err)
		defer func() { _ = os.Remove(tempFile.Name()) }()

		_, err = tempFile.WriteString(configContent)
		require.NoError(t, err)
		_ = tempFile.Close()

		cfg := MustLoad(tempFile.Name())

		assert.Equal(t, "INFO", cfg.LogLevel)
		assert.Equal(t, 2*time.Hour, cfg.IndexTTL)
		assert.Equal(t, "localhost:8080", cfg.Address)
		assert.Equal(t, "localhost:5432", cfg.DBAddress)
		assert.Equal(t, "localhost:8081", cfg.WordsAddress)
		assert.Equal(t, "nats://localhost:4223", cfg.BrokerAddress)
	})

	t.Run("config file not found", func(t *testing.T) {

	})
}

func TestConfig_StructTags(t *testing.T) {
	cfg := Config{
		LogLevel:      "DEBUG",
		IndexTTL:      time.Hour,
		Address:       "localhost:80",
		DBAddress:     "localhost:82",
		WordsAddress:  "localhost:81",
		BrokerAddress: "nats://localhost:4222",
	}

	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.Equal(t, time.Hour, cfg.IndexTTL)
	assert.Equal(t, "localhost:80", cfg.Address)
	assert.Equal(t, "localhost:82", cfg.DBAddress)
	assert.Equal(t, "localhost:81", cfg.WordsAddress)
	assert.Equal(t, "nats://localhost:4222", cfg.BrokerAddress)
}
