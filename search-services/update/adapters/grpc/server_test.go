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
	updatepb "yadro.com/course/proto/update"
	"yadro.com/course/update/core"
)

type MockUpdater struct {
	mock.Mock
}

func (m *MockUpdater) Update(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockUpdater) Stats(ctx context.Context) (core.ServiceStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(core.ServiceStats), args.Error(1)
}

func (m *MockUpdater) Status(ctx context.Context) core.ServiceStatus {
	args := m.Called(ctx)
	return args.Get(0).(core.ServiceStatus)
}

func (m *MockUpdater) Drop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestNewServer(t *testing.T) {
	updater := &MockUpdater{}
	server := NewServer(updater)

	assert.NotNil(t, server)
	assert.Equal(t, updater, server.service)
}

func TestServer_Ping(t *testing.T) {
	updater := &MockUpdater{}
	server := NewServer(updater)

	resp, err := server.Ping(context.Background(), &emptypb.Empty{})

	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestServer_Status(t *testing.T) {
	updater := &MockUpdater{}
	server := NewServer(updater)

	t.Run("idle status", func(t *testing.T) {
		updater.On("Status", mock.Anything).Return(core.StatusIdle, nil).Once()

		resp, err := server.Status(context.Background(), &emptypb.Empty{})

		assert.NoError(t, err)
		assert.Equal(t, updatepb.Status_STATUS_IDLE, resp.Status)
		updater.AssertExpectations(t)
	})

	t.Run("running status", func(t *testing.T) {
		updater.On("Status", mock.Anything).Return(core.StatusRunning, nil).Once()

		resp, err := server.Status(context.Background(), &emptypb.Empty{})

		assert.NoError(t, err)
		assert.Equal(t, updatepb.Status_STATUS_RUNNING, resp.Status)
		updater.AssertExpectations(t)
	})
}

func TestServer_Update(t *testing.T) {
	updater := &MockUpdater{}
	server := NewServer(updater)

	t.Run("success", func(t *testing.T) {
		updater.On("Update", mock.Anything).Return(nil).Once()

		resp, err := server.Update(context.Background(), &emptypb.Empty{})

		assert.NoError(t, err)
		assert.Nil(t, resp)
		updater.AssertExpectations(t)
	})

	t.Run("already exists", func(t *testing.T) {
		updater.On("Update", mock.Anything).Return(core.ErrAlreadyExists).Once()

		resp, err := server.Update(context.Background(), &emptypb.Empty{})

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
		assert.Nil(t, resp)
		updater.AssertExpectations(t)
	})

	t.Run("other error", func(t *testing.T) {
		updater.On("Update", mock.Anything).Return(errors.New("other error")).Once()

		resp, err := server.Update(context.Background(), &emptypb.Empty{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "other error")
		assert.Nil(t, resp)
		updater.AssertExpectations(t)
	})
}

func TestServer_Stats(t *testing.T) {
	updater := &MockUpdater{}
	server := NewServer(updater)

	t.Run("success", func(t *testing.T) {
		expectedStats := core.ServiceStats{
			DBStats: core.DBStats{
				WordsTotal:    100,
				WordsUnique:   50,
				ComicsFetched: 25,
			},
			ComicsTotal: 1000,
		}
		updater.On("Stats", mock.Anything).Return(expectedStats, nil).Once()

		resp, err := server.Stats(context.Background(), &emptypb.Empty{})

		assert.NoError(t, err)
		assert.Equal(t, int64(100), resp.WordsTotal)
		assert.Equal(t, int64(50), resp.WordsUnique)
		assert.Equal(t, int64(1000), resp.ComicsTotal)
		assert.Equal(t, int64(25), resp.ComicsFetched)
		updater.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		updater.On("Stats", mock.Anything).Return(core.ServiceStats{}, errors.New("stats error")).Once()

		resp, err := server.Stats(context.Background(), &emptypb.Empty{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stats error")
		assert.Nil(t, resp)
		updater.AssertExpectations(t)
	})
}

func TestServer_Drop(t *testing.T) {
	updater := &MockUpdater{}
	server := NewServer(updater)

	t.Run("success", func(t *testing.T) {
		updater.On("Drop", mock.Anything).Return(nil).Once()

		resp, err := server.Drop(context.Background(), &emptypb.Empty{})

		assert.NoError(t, err)
		assert.Nil(t, resp)
		updater.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		updater.On("Drop", mock.Anything).Return(errors.New("drop error")).Once()

		resp, err := server.Drop(context.Background(), &emptypb.Empty{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "drop error")
		assert.Nil(t, resp)
		updater.AssertExpectations(t)
	})
}
