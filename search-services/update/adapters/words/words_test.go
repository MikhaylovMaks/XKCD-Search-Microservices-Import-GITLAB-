package words

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
	"yadro.com/course/update/core"
)

type MockWordsClient struct {
	mock.Mock
}

func (m *MockWordsClient) Norm(ctx context.Context, in *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wordspb.WordsReply), args.Error(1)
}

func (m *MockWordsClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func TestNewClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("empty address", func(t *testing.T) {
		client, err := NewClient("", logger)

		// grpc.NewClient may not validate address synchronously
		// Just check that function doesn't panic
		assert.NotNil(t, client)
		assert.NoError(t, err)
	})
}

func TestClient_Norm(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockClient := &MockWordsClient{}

	client := &Client{
		log:    logger,
		client: mockClient,
	}

	t.Run("success", func(t *testing.T) {
		expectedWords := []string{"hello", "world"}
		reply := &wordspb.WordsReply{Words: expectedWords}

		mockClient.On("Norm", mock.Anything, &wordspb.WordsRequest{Phrase: "Hello World"}, mock.Anything).
			Return(reply, nil).Once()

		words, err := client.Norm(context.Background(), "Hello World")

		assert.NoError(t, err)
		assert.Equal(t, expectedWords, words)
		mockClient.AssertExpectations(t)
	})

	t.Run("resource exhausted", func(t *testing.T) {
		mockClient.On("Norm", mock.Anything, &wordspb.WordsRequest{Phrase: "long phrase"}, mock.Anything).
			Return(nil, status.Error(codes.ResourceExhausted, "too long")).Once()

		words, err := client.Norm(context.Background(), "long phrase")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, core.ErrBadArguments))
		assert.Nil(t, words)
		mockClient.AssertExpectations(t)
	})

	t.Run("other error", func(t *testing.T) {
		mockClient.On("Norm", mock.Anything, &wordspb.WordsRequest{Phrase: "test"}, mock.Anything).
			Return(nil, errors.New("grpc error")).Once()

		words, err := client.Norm(context.Background(), "test")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "grpc error")
		assert.Nil(t, words)
		mockClient.AssertExpectations(t)
	})
}

func TestClient_Ping(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockClient := &MockWordsClient{}

	client := &Client{
		log:    logger,
		client: mockClient,
	}

	t.Run("success", func(t *testing.T) {
		mockClient.On("Ping", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).
			Return(&emptypb.Empty{}, nil).Once()

		err := client.Ping(context.Background())

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockClient.On("Ping", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).
			Return(nil, errors.New("ping error")).Once()

		err := client.Ping(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ping error")
		mockClient.AssertExpectations(t)
	})
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
