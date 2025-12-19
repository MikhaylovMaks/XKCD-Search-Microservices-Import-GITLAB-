package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"yadro.com/course/api/core"
)

// MockAuthenticator is a mock implementation of Authenticator
type MockAuthenticator struct {
	mock.Mock
}

func (m *MockAuthenticator) Login(user, password string) (string, error) {
	args := m.Called(user, password)
	return args.String(0), args.Error(1)
}

// MockPinger is a mock implementation of core.Pinger
type MockPinger struct {
	mock.Mock
}

func (m *MockPinger) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockUpdater is a mock implementation of core.Updater
type MockUpdater struct {
	mock.Mock
}

func (m *MockUpdater) Update(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockUpdater) Stats(ctx context.Context) (core.UpdateStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return core.UpdateStats{}, args.Error(1)
	}
	return args.Get(0).(core.UpdateStats), args.Error(1)
}

func (m *MockUpdater) Status(ctx context.Context) (core.UpdateStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return core.StatusUpdateUnknown, args.Error(1)
	}
	return args.Get(0).(core.UpdateStatus), args.Error(1)
}

func (m *MockUpdater) Drop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockSearcher is a mock implementation of core.Searcher
type MockSearcher struct {
	mock.Mock
}

func (m *MockSearcher) Search(ctx context.Context, phrase string, limit int) ([]core.Comics, error) {
	args := m.Called(ctx, phrase, limit)
	if len(args) == 0 {
		return nil, nil
	}
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	comics, ok := result.([]core.Comics)
	if !ok {
		return nil, args.Error(1)
	}
	return comics, args.Error(1)
}

func (m *MockSearcher) SearchIndex(ctx context.Context, phrase string, limit int) ([]core.Comics, error) {
	args := m.Called(ctx, phrase, limit)
	if len(args) == 0 {
		return nil, nil
	}
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	comics, ok := result.([]core.Comics)
	if !ok {
		return nil, args.Error(1)
	}
	return comics, args.Error(1)
}

func TestEncodeReply(t *testing.T) {
	var buf bytes.Buffer

	reply := map[string]string{"status": "ok", "message": "test"}

	err := encodeReply(&buf, reply)
	require.NoError(t, err)

	var decoded map[string]string
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Equal(t, reply, decoded)

	// Check that output has proper indentation
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	assert.True(t, len(lines) > 1, "should have multiple lines due to indentation")
}

func TestNewPingHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	mockPinger1 := &MockPinger{}
	mockPinger2 := &MockPinger{}

	pingers := map[string]core.Pinger{
		"service1": mockPinger1,
		"service2": mockPinger2,
	}

	handler := NewPingHandler(logger, pingers)

	tests := []struct {
		name         string
		mockSetup    func()
		expectedBody PingResponse
	}{
		{
			name: "all services available",
			mockSetup: func() {
				mockPinger1.On("Ping", mock.Anything).Return(nil).Once()
				mockPinger2.On("Ping", mock.Anything).Return(nil).Once()
			},
			expectedBody: PingResponse{
				Replies: map[string]string{
					"service1": "ok",
					"service2": "ok",
				},
			},
		},
		{
			name: "one service unavailable",
			mockSetup: func() {
				mockPinger1.On("Ping", mock.Anything).Return(nil).Once()
				mockPinger2.On("Ping", mock.Anything).Return(errors.New("connection failed")).Once()
			},
			expectedBody: PingResponse{
				Replies: map[string]string{
					"service1": "ok",
					"service2": "unavailable",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response PingResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedBody.Replies, response.Replies)

			mockPinger1.AssertExpectations(t)
			mockPinger2.AssertExpectations(t)
		})
	}
}

func TestNewLoginHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockAuth := &MockAuthenticator{}

	handler := NewLoginHandler(logger, mockAuth)

	tests := []struct {
		name         string
		requestBody  string
		mockSetup    func()
		expectedCode int
		expectToken  bool
	}{
		{
			name:        "successful login",
			requestBody: `{"name":"admin","password":"password"}`,
			mockSetup: func() {
				mockAuth.On("Login", "admin", "password").Return("token123", nil).Once()
			},
			expectedCode: http.StatusOK,
			expectToken:  true,
		},
		{
			name:         "invalid json",
			requestBody:  `{invalid json}`,
			mockSetup:    func() {},
			expectedCode: http.StatusBadRequest,
			expectToken:  false,
		},
		{
			name:        "authentication failed",
			requestBody: `{"name":"admin","password":"wrong"}`,
			mockSetup: func() {
				mockAuth.On("Login", "admin", "wrong").Return("", errors.New("invalid credentials")).Once()
			},
			expectedCode: http.StatusUnauthorized,
			expectToken:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte(tt.requestBody)))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectToken {
				assert.NotEmpty(t, w.Body.String())
				assert.Contains(t, w.Body.String(), "token123")
			}

			mockAuth.AssertExpectations(t)
		})
	}
}

func TestNewUpdateHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockUpdater := &MockUpdater{}

	handler := NewUpdateHandler(logger, mockUpdater)

	tests := []struct {
		name         string
		mockSetup    func()
		expectedCode int
	}{
		{
			name: "successful update",
			mockSetup: func() {
				mockUpdater.On("Update", mock.Anything).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "already exists error",
			mockSetup: func() {
				mockUpdater.On("Update", mock.Anything).Return(core.ErrAlreadyExists).Once()
			},
			expectedCode: http.StatusAccepted,
		},
		{
			name: "other error",
			mockSetup: func() {
				mockUpdater.On("Update", mock.Anything).Return(errors.New("database error")).Once()
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodPost, "/update", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			mockUpdater.AssertExpectations(t)
		})
	}
}

func TestNewUpdateStatsHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockUpdater := &MockUpdater{}

	handler := NewUpdateStatsHandler(logger, mockUpdater)

	t.Run("success", func(t *testing.T) {
		stats := core.UpdateStats{
			WordsTotal:    1000,
			WordsUnique:   500,
			ComicsFetched: 100,
			ComicsTotal:   200,
		}

		mockUpdater.On("Stats", mock.Anything).Return(stats, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/stats", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response UpdateStats
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, 1000, response.WordsTotal)
		assert.Equal(t, 500, response.WordsUnique)
		assert.Equal(t, 100, response.ComicsFetched)
		assert.Equal(t, 200, response.ComicsTotal)

		mockUpdater.AssertExpectations(t)
	})

	t.Run("stats error", func(t *testing.T) {
		mockUpdater.On("Stats", mock.Anything).Return(core.UpdateStats{}, errors.New("stats error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/stats", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "stats error")

		mockUpdater.AssertExpectations(t)
	})
}

func TestNewUpdateStatusHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockUpdater := &MockUpdater{}

	handler := NewUpdateStatusHandler(logger, mockUpdater)

	t.Run("success idle", func(t *testing.T) {
		mockUpdater.On("Status", mock.Anything).Return(core.StatusUpdateIdle, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/status", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response UpdateStatus
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "idle", response.Status)

		mockUpdater.AssertExpectations(t)
	})

	t.Run("success running", func(t *testing.T) {
		mockUpdater.On("Status", mock.Anything).Return(core.StatusUpdateRunning, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/status", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response UpdateStatus
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "running", response.Status)

		mockUpdater.AssertExpectations(t)
	})

	t.Run("success unknown", func(t *testing.T) {
		mockUpdater.On("Status", mock.Anything).Return(core.StatusUpdateUnknown, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/status", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response UpdateStatus
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "unknown", response.Status)

		mockUpdater.AssertExpectations(t)
	})

	t.Run("status error", func(t *testing.T) {
		mockUpdater.On("Status", mock.Anything).Return(core.StatusUpdateUnknown, errors.New("status error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/status", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "status error")

		mockUpdater.AssertExpectations(t)
	})
}

func TestNewDropHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockUpdater := &MockUpdater{}

	handler := NewDropHandler(logger, mockUpdater)

	tests := []struct {
		name         string
		mockSetup    func()
		expectedCode int
	}{
		{
			name: "successful drop",
			mockSetup: func() {
				mockUpdater.On("Drop", mock.Anything).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "drop error",
			mockSetup: func() {
				mockUpdater.On("Drop", mock.Anything).Return(errors.New("drop failed")).Once()
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodDelete, "/drop", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			mockUpdater.AssertExpectations(t)
		})
	}
}
