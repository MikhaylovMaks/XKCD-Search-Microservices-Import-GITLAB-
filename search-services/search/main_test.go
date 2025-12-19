package main

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustMakeLogger(t *testing.T) {
	t.Run("DEBUG level", func(t *testing.T) {
		logger := mustMakeLogger("DEBUG")
		assert.NotNil(t, logger)
		assert.IsType(t, &slog.Logger{}, logger)
	})

	t.Run("INFO level", func(t *testing.T) {
		logger := mustMakeLogger("INFO")
		assert.NotNil(t, logger)
		assert.IsType(t, &slog.Logger{}, logger)
	})

	t.Run("ERROR level", func(t *testing.T) {
		logger := mustMakeLogger("ERROR")
		assert.NotNil(t, logger)
		assert.IsType(t, &slog.Logger{}, logger)
	})

	t.Run("unknown level panics", func(t *testing.T) {
		assert.Panics(t, func() {
			mustMakeLogger("UNKNOWN")
		})
	})

	t.Run("empty level panics", func(t *testing.T) {
		assert.Panics(t, func() {
			mustMakeLogger("")
		})
	})
}

func TestMustMakeLogger_LevelValues(t *testing.T) {
	// Test that different log levels are set correctly by checking the logger behavior
	// Since we can't directly check the level, we'll test that the function returns a valid logger

	testCases := []string{"DEBUG", "INFO", "ERROR"}
	for _, level := range testCases {
		t.Run("level_"+level, func(t *testing.T) {
			logger := mustMakeLogger(level)
			require.NotNil(t, logger)

			// Test that logger can be used for basic operations
			logger.Info("test message")
			logger.Debug("debug message")
			logger.Error("error message")
		})
	}
}

// Note: The main() function and run() function are difficult to unit test
// because they involve network connections, database connections, gRPC servers,
// and complex initialization. For production code, these would typically be
// tested with integration tests or the functions would be refactored to be
// more testable. This test file provides basic coverage for the testable
// parts of main.go.
