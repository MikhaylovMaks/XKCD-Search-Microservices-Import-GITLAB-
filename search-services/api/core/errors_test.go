package core

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrBadArguments(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ErrBadArguments matches",
			err:  ErrBadArguments,
			want: true,
		},
		{
			name: "different error doesn't match",
			err:  errors.New("other error"),
			want: false,
		},
		{
			name: "nil error doesn't match",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errors.Is(tt.err, ErrBadArguments)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrAlreadyExists(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ErrAlreadyExists matches",
			err:  ErrAlreadyExists,
			want: true,
		},
		{
			name: "different error doesn't match",
			err:  errors.New("other error"),
			want: false,
		},
		{
			name: "nil error doesn't match",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errors.Is(tt.err, ErrAlreadyExists)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ErrNotFound matches",
			err:  ErrNotFound,
			want: true,
		},
		{
			name: "different error doesn't match",
			err:  errors.New("other error"),
			want: false,
		},
		{
			name: "nil error doesn't match",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errors.Is(tt.err, ErrNotFound)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrBadArguments message",
			err:      ErrBadArguments,
			expected: "arguments are not acceptable",
		},
		{
			name:     "ErrAlreadyExists message",
			err:      ErrAlreadyExists,
			expected: "resource or task already exists",
		},
		{
			name:     "ErrNotFound message",
			err:      ErrNotFound,
			expected: "resource is not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	baseErr := errors.New("base error")

	tests := []struct {
		name     string
		wrapped  error
		target   error
		expected bool
	}{
		{
			name:     "ErrBadArguments wraps base error",
			wrapped:  fmt.Errorf("wrapped: %w", ErrBadArguments),
			target:   ErrBadArguments,
			expected: true,
		},
		{
			name:     "base error doesn't match ErrBadArguments",
			wrapped:  baseErr,
			target:   ErrBadArguments,
			expected: false,
		},
		{
			name:     "ErrAlreadyExists in chain",
			wrapped:  fmt.Errorf("context: %w", ErrAlreadyExists),
			target:   ErrAlreadyExists,
			expected: true,
		},
		{
			name:     "ErrNotFound in deep chain",
			wrapped:  fmt.Errorf("level1: %w", fmt.Errorf("level2: %w", ErrNotFound)),
			target:   ErrNotFound,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.wrapped, tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorTypes(t *testing.T) {
	assert.IsType(t, errors.New(""), ErrBadArguments)
	assert.IsType(t, errors.New(""), ErrAlreadyExists)
	assert.IsType(t, errors.New(""), ErrNotFound)
	assert.NotNil(t, ErrBadArguments)
	assert.NotNil(t, ErrAlreadyExists)
	assert.NotNil(t, ErrNotFound)
}

func TestErrorUniqueness(t *testing.T) {
	assert.NotEqual(t, ErrBadArguments, ErrAlreadyExists)
	assert.NotEqual(t, ErrBadArguments, ErrNotFound)
	assert.NotEqual(t, ErrAlreadyExists, ErrNotFound)
	otherErr := errors.New("different error message")
	assert.NotEqual(t, ErrBadArguments, otherErr)
	assert.NotEqual(t, ErrAlreadyExists, otherErr)
	assert.NotEqual(t, ErrNotFound, otherErr)
}
