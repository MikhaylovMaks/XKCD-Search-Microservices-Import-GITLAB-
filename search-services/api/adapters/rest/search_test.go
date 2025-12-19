package rest

import (
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

func TestNewSearchHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockSearcher := &MockSearcher{}

	handler := NewSearchHandler(logger, mockSearcher)

	tests := []struct {
		name         string
		queryParams  string
		mockSetup    func()
		expectedCode int
		expectedBody *ComicsReply
	}{
		{
			name:        "successful search with limit",
			queryParams: "?phrase=linux&limit=5",
			mockSetup: func() {
				comics := []core.Comics{
					{ID: 1, URL: "http://example.com/1", Score: 90},
					{ID: 2, URL: "http://example.com/2", Score: 85},
				}
				mockSearcher.On("Search", mock.Anything, "linux", 5).Return(comics, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectedBody: &ComicsReply{
				Comics: []Comics{
					{ID: 1, URL: "http://example.com/1", Score: 90},
					{ID: 2, URL: "http://example.com/2", Score: 85},
				},
				Total: 2,
			},
		},
		{
			name:        "successful search without limit",
			queryParams: "?phrase=linux",
			mockSetup: func() {
				comics := []core.Comics{
					{ID: 1, URL: "http://example.com/1", Score: 90},
				}
				mockSearcher.On("Search", mock.Anything, "linux", 0).Return(comics, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectedBody: &ComicsReply{
				Comics: []Comics{
					{ID: 1, URL: "http://example.com/1", Score: 90},
				},
				Total: 1,
			},
		},
		{
			name:         "no phrase provided",
			queryParams:  "?limit=5",
			mockSetup:    func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid limit",
			queryParams:  "?phrase=linux&limit=abc",
			mockSetup:    func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "negative limit",
			queryParams:  "?phrase=linux&limit=-1",
			mockSetup:    func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:        "not found error",
			queryParams: "?phrase=nonexistent",
			mockSetup: func() {
				mockSearcher.On("Search", mock.Anything, "nonexistent", 0).Return(nil, core.ErrNotFound).Once()
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:        "search error",
			queryParams: "?phrase=linux",
			mockSetup: func() {
				mockSearcher.On("Search", mock.Anything, "linux", 0).Return(nil, errors.New("database error")).Once()
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodGet, "/search"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedBody != nil {
				var response ComicsReply
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, *tt.expectedBody, response)
			}

			mockSearcher.AssertExpectations(t)
		})
	}
}

func TestNewSearchIndexHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockSearcher := &MockSearcher{}

	handler := NewSearchIndexHandler(logger, mockSearcher)

	tests := []struct {
		name         string
		queryParams  string
		mockSetup    func()
		expectedCode int
		expectedBody *ComicsReply
	}{
		{
			name:        "successful search index with limit",
			queryParams: "?phrase=linux&limit=3",
			mockSetup: func() {
				comics := []core.Comics{
					{ID: 1, URL: "http://example.com/1", Score: 95},
					{ID: 2, URL: "http://example.com/2", Score: 90},
					{ID: 3, URL: "http://example.com/3", Score: 85},
				}
				mockSearcher.On("SearchIndex", mock.Anything, "linux", 3).Return(comics, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectedBody: &ComicsReply{
				Comics: []Comics{
					{ID: 1, URL: "http://example.com/1", Score: 95},
					{ID: 2, URL: "http://example.com/2", Score: 90},
					{ID: 3, URL: "http://example.com/3", Score: 85},
				},
				Total: 3,
			},
		},
		{
			name:        "successful search index without limit",
			queryParams: "?phrase=linux",
			mockSetup: func() {
				comics := []core.Comics{
					{ID: 1, URL: "http://example.com/1", Score: 90},
				}
				mockSearcher.On("SearchIndex", mock.Anything, "linux", 0).Return(comics, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectedBody: &ComicsReply{
				Comics: []Comics{
					{ID: 1, URL: "http://example.com/1", Score: 90},
				},
				Total: 1,
			},
		},
		{
			name:         "no phrase provided",
			queryParams:  "?limit=5",
			mockSetup:    func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid limit",
			queryParams:  "?phrase=linux&limit=xyz",
			mockSetup:    func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "negative limit",
			queryParams:  "?phrase=linux&limit=-5",
			mockSetup:    func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:        "not found error",
			queryParams: "?phrase=rareterm",
			mockSetup: func() {
				mockSearcher.On("SearchIndex", mock.Anything, "rareterm", 0).Return(nil, core.ErrNotFound).Once()
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:        "search index error",
			queryParams: "?phrase=linux",
			mockSetup: func() {
				mockSearcher.On("SearchIndex", mock.Anything, "linux", 0).Return(nil, errors.New("index error")).Once()
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodGet, "/isearch"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedBody != nil {
				var response ComicsReply
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, *tt.expectedBody, response)
			}

			mockSearcher.AssertExpectations(t)
		})
	}
}
