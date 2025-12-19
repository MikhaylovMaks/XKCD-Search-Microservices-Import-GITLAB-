package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/core"
)

// MockSearcher implements core.Searcher for testing
type MockSearcher struct {
	mock.Mock
}

func (m *MockSearcher) Search(ctx context.Context, phrase string, limit int) ([]core.Comics, error) {
	args := m.Called(ctx, phrase, limit)
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

func (m *MockSearcher) BuildIndex(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSearcher) ClearIndex() {
	m.Called()
}

func TestNewServer(t *testing.T) {
	mockSearcher := &MockSearcher{}
	server := NewServer(mockSearcher)
	assert.NotNil(t, server)
	assert.Equal(t, mockSearcher, server.service)
}

func TestServer_Ping(t *testing.T) {
	mockSearcher := &MockSearcher{}
	server := NewServer(mockSearcher)

	resp, err := server.Ping(context.Background(), &emptypb.Empty{})

	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestServer_Search(t *testing.T) {
	mockSearcher := &MockSearcher{}
	server := NewServer(mockSearcher)

	tests := []struct {
		name        string
		req         *searchpb.SearchRequest
		mockResult  []core.Comics
		mockError   error
		expectError bool
		errorCode   codes.Code
	}{
		{
			name: "successful search with results",
			req:  &searchpb.SearchRequest{Phrase: "test", Limit: 5},
			mockResult: []core.Comics{
				{ID: 1, URL: "http://example.com/1", Score: 90},
				{ID: 2, URL: "http://example.com/2", Score: 85},
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "search with zero limit uses default",
			req:         &searchpb.SearchRequest{Phrase: "test", Limit: 0},
			mockResult:  []core.Comics{},
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "search not found",
			req:         &searchpb.SearchRequest{Phrase: "nonexistent", Limit: 5},
			mockResult:  nil,
			mockError:   core.ErrNotFound,
			expectError: true,
			errorCode:   codes.NotFound,
		},
		{
			name:        "search error",
			req:         &searchpb.SearchRequest{Phrase: "error", Limit: 5},
			mockResult:  nil,
			mockError:   errors.New("database error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedLimit := int(tt.req.Limit)
			if expectedLimit == 0 {
				expectedLimit = defaultLimit
			}

			mockSearcher.On("Search", mock.Anything, tt.req.Phrase, expectedLimit).
				Return(tt.mockResult, tt.mockError).Once()

			resp, err := server.Search(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorCode != 0 {
					st, ok := status.FromError(err)
					assert.True(t, ok)
					assert.Equal(t, tt.errorCode, st.Code())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Len(t, resp.Comics, len(tt.mockResult))
				for i, comic := range tt.mockResult {
					assert.Equal(t, int64(comic.ID), resp.Comics[i].Id)
					assert.Equal(t, comic.URL, resp.Comics[i].Url)
					assert.Equal(t, int64(comic.Score), resp.Comics[i].Score)
				}
			}

			mockSearcher.AssertExpectations(t)
		})
	}
}

func TestServer_SearchIndex(t *testing.T) {
	mockSearcher := &MockSearcher{}
	server := NewServer(mockSearcher)

	t.Run("successful search index", func(t *testing.T) {
		req := &searchpb.SearchRequest{Phrase: "test", Limit: 3}
		expectedResults := []core.Comics{
			{ID: 10, URL: "http://example.com/10", Score: 95},
		}

		mockSearcher.On("SearchIndex", mock.Anything, req.Phrase, int(req.Limit)).
			Return(expectedResults, nil).Once()

		resp, err := server.SearchIndex(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Comics, 1)
		assert.Equal(t, int64(10), resp.Comics[0].Id)
		assert.Equal(t, "http://example.com/10", resp.Comics[0].Url)
		assert.Equal(t, int64(95), resp.Comics[0].Score)

		mockSearcher.AssertExpectations(t)
	})
}
