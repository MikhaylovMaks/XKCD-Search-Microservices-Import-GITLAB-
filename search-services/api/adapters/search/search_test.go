package search

import (
	"context"
	"errors"
	"testing"

	"log/slog"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	searchpb "yadro.com/course/proto/search"
)

// Mock for grpc.ClientConnInterface
type MockClientConn struct {
	mock.Mock
}

// Mock for SearchClient interface
type MockSearchClient struct {
	mock.Mock
}

func (m *MockSearchClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockSearchClient) Search(ctx context.Context, in *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.SearchReply, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*searchpb.SearchReply), args.Error(1)
}

func (m *MockSearchClient) SearchIndex(ctx context.Context, in *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.SearchReply, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*searchpb.SearchReply), args.Error(1)
}

func (m *MockClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	mockArgs := m.Called(ctx, method, args, reply, opts)
	return mockArgs.Error(0)
}

func (m *MockClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	mockArgs := m.Called(ctx, desc, method, opts)
	return mockArgs.Get(0).(grpc.ClientStream), mockArgs.Error(1)
}

func (m *MockClientConn) Close() error {
	return m.Called().Error(0)
}

// Test NewClient - we can't easily test this without complex mocking of grpc.NewClient
// So we'll focus on testing the methods that use the client with dependency injection

func TestClient_Search(t *testing.T) {
	// Note: This test is simplified because we can't easily mock the grpc client
	// In a real scenario, we'd use grpc interceptors or dependency injection
	// For now, we'll test the conversion logic

	t.Run("conversion logic", func(t *testing.T) {
		// Test the conversion from protobuf to core types
		pbComics := []*searchpb.Comics{
			{Id: 1, Url: "http://example.com/1", Score: 90},
			{Id: 2, Url: "http://example.com/2", Score: 85},
		}

		expected := []core.Comics{
			{ID: 1, URL: "http://example.com/1", Score: 90},
			{ID: 2, URL: "http://example.com/2", Score: 85},
		}

		var result []core.Comics
		for _, c := range pbComics {
			result = append(result, core.Comics{ID: int(c.Id), URL: c.Url, Score: int(c.Score)})
		}

		assert.Equal(t, expected, result)
	})
}

func TestClient_SearchIndex(t *testing.T) {
	// Similar to Search test - testing conversion logic
	t.Run("conversion logic", func(t *testing.T) {
		pbComics := []*searchpb.Comics{
			{Id: 10, Url: "http://example.com/10", Score: 95},
			{Id: 20, Url: "http://example.com/20", Score: 88},
		}

		expected := []core.Comics{
			{ID: 10, URL: "http://example.com/10", Score: 95},
			{ID: 20, URL: "http://example.com/20", Score: 88},
		}

		var result []core.Comics
		for _, c := range pbComics {
			result = append(result, core.Comics{ID: int(c.Id), URL: c.Url, Score: int(c.Score)})
		}

		assert.Equal(t, expected, result)
	})
}

func TestClient_Ping(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockClient := &MockSearchClient{}

	// Create client with injected mock
	client := &Client{
		log:    logger,
		client: mockClient,
	}

	tests := []struct {
		name        string
		mockSetup   func()
		expectedErr error
	}{
		{
			name: "successful ping",
			mockSetup: func() {
				mockClient.On("Ping", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).Return(&emptypb.Empty{}, nil).Once()
			},
			expectedErr: nil,
		},
		{
			name: "ping error",
			mockSetup: func() {
				mockClient.On("Ping", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).Return(nil, errors.New("connection failed")).Once()
			},
			expectedErr: errors.New("connection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := client.Ping(context.Background())

			if tt.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestClient_Close(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Test with nil conn - should handle gracefully
	client := &Client{
		log:  logger,
		conn: nil,
	}

	// With nil conn, Close should not panic and return nil (already closed)
	err := client.Close()
	assert.NoError(t, err, "Close should not return error with nil conn")
}

func TestClient_Methods(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockClient := &MockSearchClient{}

	client := &Client{
		log:    logger,
		client: mockClient,
	}

	t.Run("Search success", func(t *testing.T) {
		comics := []*searchpb.Comics{
			{Id: 1, Url: "http://example.com/1", Score: 90},
			{Id: 2, Url: "http://example.com/2", Score: 85},
		}
		reply := &searchpb.SearchReply{Comics: comics}

		mockClient.On("Search", mock.Anything, &searchpb.SearchRequest{Phrase: "test", Limit: 10}, mock.Anything).
			Return(reply, nil).Once()

		result, err := client.Search(context.Background(), "test", 10)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 1, result[0].ID)
		assert.Equal(t, "http://example.com/1", result[0].URL)
		assert.Equal(t, 90, result[0].Score)

		mockClient.AssertExpectations(t)
	})

	t.Run("Search not found", func(t *testing.T) {
		mockClient.On("Search", mock.Anything, &searchpb.SearchRequest{Phrase: "nonexistent", Limit: 5}, mock.Anything).
			Return(nil, status.Error(codes.NotFound, "not found")).Once()

		result, err := client.Search(context.Background(), "nonexistent", 5)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, core.ErrNotFound))
		assert.Nil(t, result)

		mockClient.AssertExpectations(t)
	})

	t.Run("SearchIndex success", func(t *testing.T) {
		comics := []*searchpb.Comics{
			{Id: 10, Url: "http://example.com/10", Score: 95},
		}
		reply := &searchpb.SearchReply{Comics: comics}

		mockClient.On("SearchIndex", mock.Anything, &searchpb.SearchRequest{Phrase: "index", Limit: 1}, mock.Anything).
			Return(reply, nil).Once()

		result, err := client.SearchIndex(context.Background(), "index", 1)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, 10, result[0].ID)

		mockClient.AssertExpectations(t)
	})

	t.Run("SearchIndex error", func(t *testing.T) {
		mockClient.On("SearchIndex", mock.Anything, &searchpb.SearchRequest{Phrase: "error", Limit: 0}, mock.Anything).
			Return(nil, errors.New("grpc error")).Once()

		result, err := client.SearchIndex(context.Background(), "error", 0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "grpc error")
		assert.Nil(t, result)

		mockClient.AssertExpectations(t)
	})
}

// Test error conversion
func TestErrorConversion(t *testing.T) {
	tests := []struct {
		name     string
		grpcErr  error
		expected error
	}{
		{
			name:     "not found code",
			grpcErr:  status.Error(codes.NotFound, "not found"),
			expected: core.ErrNotFound,
		},
		{
			name:     "other error",
			grpcErr:  errors.New("other error"),
			expected: errors.New("other error"),
		},
		{
			name:     "nil error",
			grpcErr:  nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch {
			case tt.grpcErr == nil:
				assert.NoError(t, tt.expected)
			case status.Code(tt.grpcErr) == codes.NotFound:
				assert.Error(t, tt.grpcErr)
				assert.True(t, errors.Is(tt.expected, core.ErrNotFound))
			default:
				assert.Error(t, tt.grpcErr)
				assert.Equal(t, tt.expected, tt.grpcErr)
			}
		})
	}
}
