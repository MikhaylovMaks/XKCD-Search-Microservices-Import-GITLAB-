package words

import (
	"context"
	"errors"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	wordspb "yadro.com/course/proto/words"
)

type mockWordsClient struct {
	mock.Mock
}

func (m *mockWordsClient) Norm(ctx context.Context, req *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*wordspb.WordsReply), args.Error(1)
}

func (m *mockWordsClient) Ping(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func TestNewClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, nil))

	t.Run("connection failure", func(t *testing.T) {
		client, err := NewClient("invalid:address", logger)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.conn)
	})
}

func TestClient_Close(t *testing.T) {
	t.Run("close with nil connection", func(t *testing.T) {
		client := &Client{}
		err := client.Close()
		assert.Error(t, err)
	})

	t.Run("close with logger", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(nil, nil))
		client := &Client{log: logger}
		err := client.Close()
		assert.Error(t, err)
	})
}

func TestClient_Norm(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	ctx := context.Background()

	t.Run("successful normalization", func(t *testing.T) {
		mockClient := &mockWordsClient{}
		client := &Client{
			log:    logger,
			client: mockClient,
		}

		expectedWords := []string{"hello", "world"}
		mockClient.On("Norm", ctx, &wordspb.WordsRequest{Phrase: "Hello World"}).Return(
			&wordspb.WordsReply{Words: expectedWords}, nil).Once()

		result, err := client.Norm(ctx, "Hello World")

		assert.NoError(t, err)
		assert.Equal(t, expectedWords, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("resource exhausted error", func(t *testing.T) {
		mockClient := &mockWordsClient{}
		client := &Client{
			log:    logger,
			client: mockClient,
		}

		resourceExhaustedErr := status.Error(codes.ResourceExhausted, "rate limited")
		mockClient.On("Norm", ctx, &wordspb.WordsRequest{Phrase: "test"}).Return(
			(*wordspb.WordsReply)(nil), resourceExhaustedErr).Once()

		result, err := client.Norm(ctx, "test")

		assert.Error(t, err)
		assert.Equal(t, core.ErrBadArguments, err)
		assert.Nil(t, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("other gRPC error", func(t *testing.T) {
		mockClient := &mockWordsClient{}
		client := &Client{
			log:    logger,
			client: mockClient,
		}

		grpcErr := status.Error(codes.Internal, "internal error")
		mockClient.On("Norm", ctx, &wordspb.WordsRequest{Phrase: "test"}).Return(
			(*wordspb.WordsReply)(nil), grpcErr).Once()

		result, err := client.Norm(ctx, "test")

		assert.Error(t, err)
		assert.Equal(t, grpcErr, err)
		assert.Nil(t, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("generic error", func(t *testing.T) {
		mockClient := &mockWordsClient{}
		client := &Client{
			log:    logger,
			client: mockClient,
		}

		genericErr := errors.New("connection error")
		mockClient.On("Norm", ctx, &wordspb.WordsRequest{Phrase: "test"}).Return(
			(*wordspb.WordsReply)(nil), genericErr).Once()

		result, err := client.Norm(ctx, "test")

		assert.Error(t, err)
		assert.Equal(t, genericErr, err)
		assert.Nil(t, result)
		mockClient.AssertExpectations(t)
	})
}

func TestClient_Ping(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	ctx := context.Background()

	t.Run("successful ping", func(t *testing.T) {
		mockClient := &mockWordsClient{}
		client := &Client{
			log:    logger,
			client: mockClient,
		}

		mockClient.On("Ping", ctx, &emptypb.Empty{}).Return(
			&emptypb.Empty{}, nil).Once()

		err := client.Ping(ctx)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("service unavailable", func(t *testing.T) {
		mockClient := &mockWordsClient{}
		client := &Client{
			log:    logger,
			client: mockClient,
		}

		unavailableErr := status.Error(codes.Unavailable, "service unavailable")
		mockClient.On("Ping", ctx, &emptypb.Empty{}).Return(
			(*emptypb.Empty)(nil), unavailableErr).Once()

		err := client.Ping(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service unavailable")
		mockClient.AssertExpectations(t)
	})

	t.Run("other gRPC error", func(t *testing.T) {
		mockClient := &mockWordsClient{}
		client := &Client{
			log:    logger,
			client: mockClient,
		}

		grpcErr := status.Error(codes.Internal, "internal error")
		mockClient.On("Ping", ctx, &emptypb.Empty{}).Return(
			(*emptypb.Empty)(nil), grpcErr).Once()

		err := client.Ping(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ping failed")
		mockClient.AssertExpectations(t)
	})

	t.Run("generic error", func(t *testing.T) {
		mockClient := &mockWordsClient{}
		client := &Client{
			log:    logger,
			client: mockClient,
		}

		genericErr := errors.New("connection error")
		mockClient.On("Ping", ctx, &emptypb.Empty{}).Return(
			(*emptypb.Empty)(nil), genericErr).Once()

		err := client.Ping(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ping failed")
		mockClient.AssertExpectations(t)
	})
}
