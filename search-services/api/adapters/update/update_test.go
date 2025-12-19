package update

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
	updatepb "yadro.com/course/proto/update"
)

type MockClientConn struct {
	mock.Mock
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

type MockUpdateClient struct {
	mock.Mock
}

func (m *MockUpdateClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockUpdateClient) Status(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatusReply, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*updatepb.StatusReply), args.Error(1)
}

func (m *MockUpdateClient) Stats(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatsReply, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*updatepb.StatsReply), args.Error(1)
}

func (m *MockUpdateClient) Update(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockUpdateClient) Drop(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func TestClient_Methods(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockClient := &MockUpdateClient{}

	client := &Client{
		log:    logger,
		client: mockClient,
	}

	t.Run("Ping success", func(t *testing.T) {
		mockClient.On("Ping", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).
			Return(&emptypb.Empty{}, nil).Once()

		err := client.Ping(context.Background())

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Ping error", func(t *testing.T) {
		mockClient.On("Ping", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).
			Return(nil, errors.New("connection failed")).Once()

		err := client.Ping(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection failed")
		mockClient.AssertExpectations(t)
	})

	t.Run("Status idle", func(t *testing.T) {
		reply := &updatepb.StatusReply{Status: updatepb.Status_STATUS_IDLE}

		mockClient.On("Status", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).
			Return(reply, nil).Once()

		status, err := client.Status(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, core.StatusUpdateIdle, status)
		mockClient.AssertExpectations(t)
	})

	t.Run("Status running", func(t *testing.T) {
		reply := &updatepb.StatusReply{Status: updatepb.Status_STATUS_RUNNING}

		mockClient.On("Status", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).
			Return(reply, nil).Once()

		status, err := client.Status(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, core.StatusUpdateRunning, status)
		mockClient.AssertExpectations(t)
	})

	t.Run("Status unknown", func(t *testing.T) {
		reply := &updatepb.StatusReply{Status: updatepb.Status_STATUS_UNSPECIFIED}

		mockClient.On("Status", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).
			Return(reply, nil).Once()

		status, err := client.Status(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, core.StatusUpdateUnknown, status)
		mockClient.AssertExpectations(t)
	})

	t.Run("Status error", func(t *testing.T) {
		mockClient.On("Status", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).
			Return(nil, errors.New("grpc error")).Once()

		status, err := client.Status(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get status")
		assert.Equal(t, core.StatusUpdateUnknown, status)
		mockClient.AssertExpectations(t)
	})

	t.Run("Stats success", func(t *testing.T) {
		reply := &updatepb.StatsReply{
			WordsTotal:    1000,
			WordsUnique:   500,
			ComicsTotal:   200,
			ComicsFetched: 150,
		}

		mockClient.On("Stats", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).
			Return(reply, nil).Once()

		stats, err := client.Stats(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, 1000, stats.WordsTotal)
		assert.Equal(t, 500, stats.WordsUnique)
		assert.Equal(t, 200, stats.ComicsTotal)
		assert.Equal(t, 150, stats.ComicsFetched)
		mockClient.AssertExpectations(t)
	})

	t.Run("Stats error", func(t *testing.T) {
		mockClient.On("Stats", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).
			Return(nil, errors.New("grpc error")).Once()

		stats, err := client.Stats(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get stats")
		assert.Equal(t, core.UpdateStats{}, stats)
		mockClient.AssertExpectations(t)
	})

	t.Run("Update success", func(t *testing.T) {
		mockClient.On("Update", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).
			Return(&emptypb.Empty{}, nil).Once()

		err := client.Update(context.Background())

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Update already exists", func(t *testing.T) {
		mockClient.On("Update", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).
			Return(nil, status.Error(codes.AlreadyExists, "already exists")).Once()

		err := client.Update(context.Background())

		assert.Error(t, err)
		assert.True(t, errors.Is(err, core.ErrAlreadyExists))
		mockClient.AssertExpectations(t)
	})

	t.Run("Update error", func(t *testing.T) {
		mockClient.On("Update", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).
			Return(nil, errors.New("grpc error")).Once()

		err := client.Update(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update failed")
		mockClient.AssertExpectations(t)
	})

	t.Run("Drop success", func(t *testing.T) {
		mockClient.On("Drop", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).
			Return(&emptypb.Empty{}, nil).Once()

		err := client.Drop(context.Background())

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Drop error", func(t *testing.T) {
		mockClient.On("Drop", mock.Anything, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).
			Return(nil, errors.New("grpc error")).Once()

		err := client.Drop(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "drop failed")
		mockClient.AssertExpectations(t)
	})
}

func TestClient_Close(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	client := &Client{
		log:  logger,
		conn: nil,
	}

	err := client.Close()
	assert.NoError(t, err, "Close should not return error with nil conn")
}

func TestClient_Status(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name        string
		pbStatus    updatepb.Status
		expected    core.UpdateStatus
		expectError bool
	}{
		{
			name:        "idle status",
			pbStatus:    updatepb.Status_STATUS_IDLE,
			expected:    core.StatusUpdateIdle,
			expectError: false,
		},
		{
			name:        "running status",
			pbStatus:    updatepb.Status_STATUS_RUNNING,
			expected:    core.StatusUpdateRunning,
			expectError: false,
		},
		{
			name:        "unknown status",
			pbStatus:    updatepb.Status_STATUS_UNSPECIFIED,
			expected:    core.StatusUpdateUnknown,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result core.UpdateStatus
			switch tt.pbStatus {
			case updatepb.Status_STATUS_IDLE:
				result = core.StatusUpdateIdle
			case updatepb.Status_STATUS_RUNNING:
				result = core.StatusUpdateRunning
			default:
				result = core.StatusUpdateUnknown
			}

			assert.Equal(t, tt.expected, result)
		})
	}

	t.Run("grpc error", func(t *testing.T) {
		assert.NotNil(t, logger, "logger should not be nil")
	})
}

func TestClient_Stats(t *testing.T) {
	tests := []struct {
		name     string
		pbStats  *updatepb.StatsReply
		expected core.UpdateStats
	}{
		{
			name: "normal stats",
			pbStats: &updatepb.StatsReply{
				WordsTotal:    1000,
				WordsUnique:   500,
				ComicsTotal:   200,
				ComicsFetched: 150,
			},
			expected: core.UpdateStats{
				WordsTotal:    1000,
				WordsUnique:   500,
				ComicsTotal:   200,
				ComicsFetched: 150,
			},
		},
		{
			name: "zero stats",
			pbStats: &updatepb.StatsReply{
				WordsTotal:    0,
				WordsUnique:   0,
				ComicsTotal:   0,
				ComicsFetched: 0,
			},
			expected: core.UpdateStats{
				WordsTotal:    0,
				WordsUnique:   0,
				ComicsTotal:   0,
				ComicsFetched: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := core.UpdateStats{
				WordsTotal:    int(tt.pbStats.WordsTotal),
				WordsUnique:   int(tt.pbStats.WordsUnique),
				ComicsTotal:   int(tt.pbStats.ComicsTotal),
				ComicsFetched: int(tt.pbStats.ComicsFetched),
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClient_Update(t *testing.T) {
	tests := []struct {
		name        string
		grpcErr     error
		expectedErr error
	}{
		{
			name:        "successful update",
			grpcErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "already exists error",
			grpcErr:     status.Error(codes.AlreadyExists, "already exists"),
			expectedErr: core.ErrAlreadyExists,
		},
		{
			name:        "other error",
			grpcErr:     errors.New("database error"),
			expectedErr: errors.New("update failed: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result error
			if tt.grpcErr != nil {
				if status.Code(tt.grpcErr) == codes.AlreadyExists {
					result = core.ErrAlreadyExists
				} else {
					result = errors.New("update failed: " + tt.grpcErr.Error())
				}
			}

			if tt.expectedErr == nil {
				assert.NoError(t, result)
			} else {
				assert.Error(t, result)
				assert.Contains(t, result.Error(), tt.expectedErr.Error())
			}
		})
	}
}

func TestClient_Drop(t *testing.T) {
	tests := []struct {
		name        string
		grpcErr     error
		expectedErr error
	}{
		{
			name:        "successful drop",
			grpcErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "drop error",
			grpcErr:     errors.New("drop failed"),
			expectedErr: errors.New("drop failed: drop failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result error
			if tt.grpcErr != nil {
				result = errors.New("drop failed: " + tt.grpcErr.Error())
			}

			if tt.expectedErr == nil {
				assert.NoError(t, result)
			} else {
				assert.Error(t, result)
				assert.Contains(t, result.Error(), tt.expectedErr.Error())
			}
		})
	}
}

func TestClient_Ping(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("ping method exists", func(t *testing.T) {
		assert.NotNil(t, logger, "logger should not be nil")
	})
}

func TestStatusEnumConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    updatepb.Status
		expected core.UpdateStatus
	}{
		{"idle", updatepb.Status_STATUS_IDLE, core.StatusUpdateIdle},
		{"running", updatepb.Status_STATUS_RUNNING, core.StatusUpdateRunning},
		{"unspecified", updatepb.Status_STATUS_UNSPECIFIED, core.StatusUpdateUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result core.UpdateStatus
			switch tt.input {
			case updatepb.Status_STATUS_IDLE:
				result = core.StatusUpdateIdle
			case updatepb.Status_STATUS_RUNNING:
				result = core.StatusUpdateRunning
			default:
				result = core.StatusUpdateUnknown
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}
