package aaa

import (
	"os"
	"testing"
	"time"

	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Set up environment variables
	_ = os.Setenv("ADMIN_USER", "testuser")
	_ = os.Setenv("ADMIN_PASSWORD", "testpass")
	defer func() {
		_ = os.Unsetenv("ADMIN_USER")
		_ = os.Unsetenv("ADMIN_PASSWORD")
	}()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	tokenTTL := 30 * time.Minute

	aaa, err := New(tokenTTL, logger)
	require.NoError(t, err)
	assert.NotNil(t, aaa)
	assert.Equal(t, tokenTTL, aaa.tokenTTL)
	assert.NotNil(t, aaa.users)
}

func TestNew_MissingUserEnv(t *testing.T) {
	// Unset ADMIN_USER
	_ = os.Unsetenv("ADMIN_USER")
	_ = os.Setenv("ADMIN_PASSWORD", "testpass")
	defer func() {
		_ = os.Unsetenv("ADMIN_PASSWORD")
	}()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	tokenTTL := 30 * time.Minute

	aaa, err := New(tokenTTL, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not get admin user")
	assert.Equal(t, AAA{}, aaa)
}

func TestNew_MissingPasswordEnv(t *testing.T) {
	// Unset ADMIN_PASSWORD
	_ = os.Setenv("ADMIN_USER", "testuser")
	_ = os.Unsetenv("ADMIN_PASSWORD")
	defer func() { _ = os.Unsetenv("ADMIN_USER") }()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	tokenTTL := 30 * time.Minute

	aaa, err := New(tokenTTL, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not get admin password")
	assert.Equal(t, AAA{}, aaa)
}

func TestAAA_Login(t *testing.T) {
	// Set up environment variables
	_ = os.Setenv("ADMIN_USER", "testuser")
	_ = os.Setenv("ADMIN_PASSWORD", "testpass")
	defer func() {
		_ = os.Unsetenv("ADMIN_USER")
		_ = os.Unsetenv("ADMIN_PASSWORD")
	}()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	tokenTTL := 30 * time.Minute

	aaa, err := New(tokenTTL, logger)
	require.NoError(t, err)

	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "successful login",
			username: "testuser",
			password: "testpass",
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			password: "testpass",
			wantErr:  true,
			errMsg:   "empty user",
		},
		{
			name:     "unknown user",
			username: "unknown",
			password: "testpass",
			wantErr:  true,
			errMsg:   "unknown user",
		},
		{
			name:     "wrong password",
			username: "testuser",
			password: "wrongpass",
			wantErr:  true,
			errMsg:   "wrong password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := aaa.Login(tt.username, tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestAAA_Verify(t *testing.T) {
	// Set up environment variables
	_ = os.Setenv("ADMIN_USER", "testuser")
	_ = os.Setenv("ADMIN_PASSWORD", "testpass")
	defer func() {
		_ = os.Unsetenv("ADMIN_USER")
		_ = os.Unsetenv("ADMIN_PASSWORD")
	}()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	tokenTTL := 30 * time.Minute

	aaa, err := New(tokenTTL, logger)
	require.NoError(t, err)

	// Get a valid token
	validToken, err := aaa.Login("testuser", "testpass")
	require.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "invalid token",
			token:   "invalid.jwt.token",
			wantErr: true,
			errMsg:  "cannot parse token",
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
			errMsg:  "cannot parse token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := aaa.Verify(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
