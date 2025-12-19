package closers

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockCloser is a test implementation of io.Closer
type mockCloser struct {
	closeFunc func() error
}

func (m *mockCloser) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestCloseOrLog(t *testing.T) {
	tests := []struct {
		name      string
		closeErr  error
		expectLog bool
	}{
		{
			name:      "successful close",
			closeErr:  nil,
			expectLog: false,
		},
		{
			name:      "close with error",
			closeErr:  io.ErrUnexpectedEOF,
			expectLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{}))

			closer := &mockCloser{
				closeFunc: func() error {
					return tt.closeErr
				},
			}

			// Should not panic
			CloseOrLog(closer, logger)

			if tt.expectLog {
				assert.Contains(t, buf.String(), "close failed")
				assert.Contains(t, buf.String(), "error")
			} else {
				assert.Empty(t, buf.String())
			}
		})
	}
}

func TestCloseOrPanic(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		closer := &mockCloser{
			closeFunc: func() error {
				return nil
			},
		}

		// Should not panic
		assert.NotPanics(t, func() {
			CloseOrPanic(closer)
		})
	})

	t.Run("close with error", func(t *testing.T) {
		closer := &mockCloser{
			closeFunc: func() error {
				return io.ErrUnexpectedEOF
			},
		}

		// Should panic
		assert.Panics(t, func() {
			CloseOrPanic(closer)
		})
	})
}
