package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTokenVerifier is a mock implementation of TokenVerifier
type MockTokenVerifier struct {
	mock.Mock
}

func (m *MockTokenVerifier) Verify(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func TestAuth(t *testing.T) {
	mockVerifier := &MockTokenVerifier{}

	// Create a test handler that will be wrapped by Auth middleware
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	authMiddleware := Auth(testHandler, mockVerifier)

	tests := []struct {
		name           string
		authHeader     string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "valid token",
			authHeader: "Token valid-token",
			mockSetup: func() {
				mockVerifier.On("Verify", "valid-token").Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "bad authorization header\n",
		},
		{
			name:           "invalid header format - no Token prefix",
			authHeader:     "Bearer valid-token",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "bad authorization header\n",
		},
		{
			name:           "invalid header format - only Token",
			authHeader:     "Token",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "bad authorization header\n",
		},
		{
			name:       "invalid token",
			authHeader: "Token invalid-token",
			mockSetup: func() {
				mockVerifier.On("Verify", "invalid-token").Return(assert.AnError).Once()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "not authorized\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			authMiddleware.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}

			mockVerifier.AssertExpectations(t)
		})
	}
}
