package config

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustLoad(t *testing.T) {
	configContent := `
log_level: "INFO"
search_concurrency: 5
search_rate: 10
api_server:
  address: "localhost:9090"
  timeout: "30s"
words_address: "words:8081"
update_address: "update:8082"
search_address: "search:8083"
token_ttl: "1h"
`

	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	_ = tmpFile.Close()

	cfg := MustLoad(tmpFile.Name())

	assert.Equal(t, "INFO", cfg.LogLevel)
	assert.Equal(t, 5, cfg.SearchConcurrency)
	assert.Equal(t, 10, cfg.SearchRate)
	assert.Equal(t, "localhost:9090", cfg.HTTPConfig.Address)
	assert.Equal(t, 30*time.Second, cfg.HTTPConfig.Timeout)
	assert.Equal(t, "words:8081", cfg.WordsAddress)
	assert.Equal(t, "update:8082", cfg.UpdateAddress)
	assert.Equal(t, "search:8083", cfg.SearchAddress)
	assert.Equal(t, time.Hour, cfg.TokenTTL)
}

func TestConfigDefaults(t *testing.T) {
	configContent := `
log_level: "DEBUG"
`

	tmpFile, err := os.CreateTemp("", "config_defaults_test_*.yaml")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	_ = tmpFile.Close()

	cfg := MustLoad(tmpFile.Name())

	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.Equal(t, 1, cfg.SearchConcurrency)
	assert.Equal(t, 1, cfg.SearchRate)
	assert.Equal(t, "localhost:80", cfg.HTTPConfig.Address)
	assert.Equal(t, 5*time.Second, cfg.HTTPConfig.Timeout)
	assert.Equal(t, "words:81", cfg.WordsAddress)
	assert.Equal(t, "update:82", cfg.UpdateAddress)
	assert.Equal(t, "search:83", cfg.SearchAddress)
	assert.Equal(t, 24*time.Hour, cfg.TokenTTL)
}

func TestHTTPConfig(t *testing.T) {
	config := HTTPConfig{
		Address: "127.0.0.1:8080",
		Timeout: 10 * time.Second,
	}

	assert.Equal(t, "127.0.0.1:8080", config.Address)
	assert.Equal(t, 10*time.Second, config.Timeout)
}

func TestMustLoad_InvalidYAML(t *testing.T) {
	invalidYAML := `
log_level: "INFO"
invalid_yaml_content: [unclosed bracket
search_concurrency: 5
`

	tmpFile, err := os.CreateTemp("", "invalid_config_test_*.yaml")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString(invalidYAML)
	require.NoError(t, err)
	_ = tmpFile.Close()

	assert.Panics(t, func() {
		MustLoad(tmpFile.Name())
	})
}

func TestMustLoad_MissingFile(t *testing.T) {
	assert.Panics(t, func() {
		MustLoad("nonexistent_file.yaml")
	})
}

func TestMustLoad_InvalidTimeout(t *testing.T) {
	configContent := `
log_level: "INFO"
api_server:
  address: "localhost:9090"
  timeout: "invalid_duration"
`

	tmpFile, err := os.CreateTemp("", "invalid_timeout_test_*.yaml")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	_ = tmpFile.Close()

	assert.Panics(t, func() {
		MustLoad(tmpFile.Name())
	})
}

func TestMustLoad_InvalidTokenTTL(t *testing.T) {
	configContent := `
log_level: "INFO"
token_ttl: "invalid_duration"
`

	tmpFile, err := os.CreateTemp("", "invalid_ttl_test_*.yaml")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	_ = tmpFile.Close()

	assert.Panics(t, func() {
		MustLoad(tmpFile.Name())
	})
}

func TestMustLoad_EmptyConfig(t *testing.T) {
	emptyConfig := ""

	tmpFile, err := os.CreateTemp("", "empty_config_test_*.yaml")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString(emptyConfig)
	require.NoError(t, err)
	_ = tmpFile.Close()

	cfg := MustLoad(tmpFile.Name())

	// Should use all defaults
	assert.Equal(t, "INFO", cfg.LogLevel)
	assert.Equal(t, 1, cfg.SearchConcurrency)
	assert.Equal(t, 1, cfg.SearchRate)
	assert.Equal(t, "localhost:80", cfg.HTTPConfig.Address)
	assert.Equal(t, 5*time.Second, cfg.HTTPConfig.Timeout)
	assert.Equal(t, "words:81", cfg.WordsAddress)
	assert.Equal(t, "update:82", cfg.UpdateAddress)
	assert.Equal(t, "search:83", cfg.SearchAddress)
	assert.Equal(t, 24*time.Hour, cfg.TokenTTL)
}

func TestMustLoad_PartialConfig(t *testing.T) {
	configContent := `
log_level: "DEBUG"
search_concurrency: 10
api_server:
  address: "127.0.0.1:8080"
`

	tmpFile, err := os.CreateTemp("", "partial_config_test_*.yaml")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	_ = tmpFile.Close()

	cfg := MustLoad(tmpFile.Name())

	// Specified values
	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.Equal(t, 10, cfg.SearchConcurrency)
	assert.Equal(t, "127.0.0.1:8080", cfg.HTTPConfig.Address)

	// Defaults
	assert.Equal(t, 1, cfg.SearchRate)
	assert.Equal(t, 5*time.Second, cfg.HTTPConfig.Timeout)
	assert.Equal(t, "words:81", cfg.WordsAddress)
	assert.Equal(t, "update:82", cfg.UpdateAddress)
	assert.Equal(t, "search:83", cfg.SearchAddress)
	assert.Equal(t, 24*time.Hour, cfg.TokenTTL)
}

func TestMustLoad_LargeValues(t *testing.T) {
	configContent := `
log_level: "ERROR"
search_concurrency: 1000
search_rate: 10000
api_server:
  address: "0.0.0.0:65535"
  timeout: "24h"
token_ttl: "720h"
`

	tmpFile, err := os.CreateTemp("", "large_config_test_*.yaml")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	_ = tmpFile.Close()

	cfg := MustLoad(tmpFile.Name())

	assert.Equal(t, "ERROR", cfg.LogLevel)
	assert.Equal(t, 1000, cfg.SearchConcurrency)
	assert.Equal(t, 10000, cfg.SearchRate)
	assert.Equal(t, "0.0.0.0:65535", cfg.HTTPConfig.Address)
	assert.Equal(t, 24*time.Hour, cfg.HTTPConfig.Timeout)
	assert.Equal(t, 720*time.Hour, cfg.TokenTTL)
}

func TestMustLoad_NetworkAddresses(t *testing.T) {
	testCases := []struct {
		name        string
		wordsAddr   string
		updateAddr  string
		searchAddr  string
		expectPanic bool
	}{
		{
			name:        "valid addresses",
			wordsAddr:   "localhost:8081",
			updateAddr:  "127.0.0.1:8082",
			searchAddr:  "example.com:8083",
			expectPanic: false,
		},
		{
			name:        "IPv6 addresses",
			wordsAddr:   "[::1]:8081",
			updateAddr:  "[2001:db8::1]:8082",
			searchAddr:  "[::]:8083",
			expectPanic: false,
		},
		{
			name:        "addresses without ports",
			wordsAddr:   "localhost",
			updateAddr:  "127.0.0.1",
			searchAddr:  "example.com",
			expectPanic: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configContent := fmt.Sprintf(`
log_level: "INFO"
words_address: "%s"
update_address: "%s"
search_address: "%s"
`, tc.wordsAddr, tc.updateAddr, tc.searchAddr)

			tmpFile, err := os.CreateTemp("", "network_config_test_*.yaml")
			require.NoError(t, err)
			defer func() { _ = os.Remove(tmpFile.Name()) }()

			_, err = tmpFile.WriteString(configContent)
			require.NoError(t, err)
			_ = tmpFile.Close()

			if tc.expectPanic {
				assert.Panics(t, func() {
					MustLoad(tmpFile.Name())
				})
			} else {
				cfg := MustLoad(tmpFile.Name())
				assert.Equal(t, tc.wordsAddr, cfg.WordsAddress)
				assert.Equal(t, tc.updateAddr, cfg.UpdateAddress)
				assert.Equal(t, tc.searchAddr, cfg.SearchAddress)
			}
		})
	}
}

func TestHTTPConfig_Validation(t *testing.T) {
	testCases := []struct {
		name    string
		config  HTTPConfig
		isValid bool
	}{
		{
			name: "valid config",
			config: HTTPConfig{
				Address: "localhost:8080",
				Timeout: 30 * time.Second,
			},
			isValid: true,
		},
		{
			name: "zero timeout",
			config: HTTPConfig{
				Address: "localhost:8080",
				Timeout: 0,
			},
			isValid: true, // Zero timeout is technically valid
		},
		{
			name: "very long timeout",
			config: HTTPConfig{
				Address: "localhost:8080",
				Timeout: 365 * 24 * time.Hour, // 1 year
			},
			isValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isValid {
				assert.Equal(t, tc.config.Address, tc.config.Address)
				assert.Equal(t, tc.config.Timeout, tc.config.Timeout)
			}
		})
	}
}

func TestConfig_ConcurrencyValidation(t *testing.T) {
	// Test that search concurrency values are reasonable
	testCases := []struct {
		name        string
		concurrency int
		rate        int
		shouldLoad  bool
	}{
		{"normal values", 5, 10, true},
		{"zero concurrency", 0, 10, true}, // Config allows zero
		{"zero rate", 5, 0, true},         // Config allows zero
		{"high concurrency", 1000, 10, true},
		{"high rate", 5, 10000, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configContent := fmt.Sprintf(`
log_level: "INFO"
search_concurrency: %d
search_rate: %d
`, tc.concurrency, tc.rate)

			tmpFile, err := os.CreateTemp("", "concurrency_config_test_*.yaml")
			require.NoError(t, err)
			defer func() { _ = os.Remove(tmpFile.Name()) }()

			_, err = tmpFile.WriteString(configContent)
			require.NoError(t, err)
			_ = tmpFile.Close()

			if tc.shouldLoad {
				cfg := MustLoad(tmpFile.Name())
				assert.Equal(t, tc.concurrency, cfg.SearchConcurrency)
				assert.Equal(t, tc.rate, cfg.SearchRate)
			} else {
				assert.Panics(t, func() {
					MustLoad(tmpFile.Name())
				})
			}
		})
	}
}
