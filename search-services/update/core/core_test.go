package core

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("valid concurrency", func(t *testing.T) {
		mockDB := &mockDB{}
		mockXKCD := &mockXKCD{}
		mockWords := &mockWords{}
		mockBroker := &mockBroker{}

		service, err := NewService(logger, mockDB, mockXKCD, mockWords, mockBroker, 2)

		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, logger, service.log)
		assert.Equal(t, mockDB, service.db)
		assert.Equal(t, mockXKCD, service.xkcd)
		assert.Equal(t, mockWords, service.words)
		assert.Equal(t, mockBroker, service.broker)
		assert.Equal(t, 2, service.concurrency)
	})

	t.Run("invalid concurrency", func(t *testing.T) {
		mockDB := &mockDB{}
		mockXKCD := &mockXKCD{}
		mockWords := &mockWords{}
		mockBroker := &mockBroker{}

		service, err := NewService(logger, mockDB, mockXKCD, mockWords, mockBroker, 0)

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "wrong concurrency specified")
	})
}

func TestService_Status(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockDB := &mockDB{}
	mockXKCD := &mockXKCD{}
	mockWords := &mockWords{}
	mockBroker := &mockBroker{}

	service, _ := NewService(logger, mockDB, mockXKCD, mockWords, mockBroker, 1)

	t.Run("idle status", func(t *testing.T) {
		// Initially should be idle
		status := service.Status(context.Background())
		assert.Equal(t, StatusIdle, status)
	})

	t.Run("running status", func(t *testing.T) {
		// Set inProgress to true
		service.inProgress.Store(true)
		defer service.inProgress.Store(false) // cleanup

		status := service.Status(context.Background())
		assert.Equal(t, StatusRunning, status)
	})
}

func TestService_Stats(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockDB := &mockDB{}
	mockXKCD := &mockXKCD{}
	mockWords := &mockWords{}
	mockBroker := &mockBroker{}
	ctx := context.Background()

	service, _ := NewService(logger, mockDB, mockXKCD, mockWords, mockBroker, 1)

	t.Run("successful stats", func(t *testing.T) {
		expectedDBStats := DBStats{
			WordsTotal:    1000,
			WordsUnique:   500,
			ComicsFetched: 200,
		}
		expectedLastID := 250

		mockDB.On("Stats", ctx).Return(expectedDBStats, nil).Once()
		mockXKCD.On("LastID", ctx).Return(expectedLastID, nil).Once()

		stats, err := service.Stats(ctx)

		assert.NoError(t, err)
		assert.Equal(t, expectedDBStats, stats.DBStats)
		assert.Equal(t, expectedLastID, stats.ComicsTotal)

		mockDB.AssertExpectations(t)
		mockXKCD.AssertExpectations(t)
	})

	t.Run("db stats error", func(t *testing.T) {
		expectedErr := assert.AnError
		mockDB.On("Stats", ctx).Return(DBStats{}, expectedErr).Once()

		stats, err := service.Stats(ctx)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, ServiceStats{}, stats)

		mockDB.AssertExpectations(t)
	})

	t.Run("xkcd last id error", func(t *testing.T) {
		expectedDBStats := DBStats{WordsTotal: 100}
		expectedErr := assert.AnError

		mockDB.On("Stats", ctx).Return(expectedDBStats, nil).Once()
		mockXKCD.On("LastID", ctx).Return(0, expectedErr).Once()

		stats, err := service.Stats(ctx)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, ServiceStats{}, stats)

		mockDB.AssertExpectations(t)
		mockXKCD.AssertExpectations(t)
	})
}

func TestService_Drop(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockDB := &mockDB{}
	mockXKCD := &mockXKCD{}
	mockWords := &mockWords{}
	mockBroker := &mockBroker{}
	ctx := context.Background()

	service, _ := NewService(logger, mockDB, mockXKCD, mockWords, mockBroker, 1)

	t.Run("successful drop with broker", func(t *testing.T) {
		mockDB.On("Drop", ctx).Return(nil).Once()
		mockBroker.On("Publish", "xkcd.db.dropped", []byte("XKCD DB has been dropped")).Return(nil).Once()

		err := service.Drop(ctx)

		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
		mockBroker.AssertExpectations(t)
	})

	t.Run("successful drop without broker", func(t *testing.T) {
		// Create service without broker
		serviceNoBroker, _ := NewService(logger, mockDB, mockXKCD, mockWords, nil, 1)

		mockDB.On("Drop", ctx).Return(nil).Once()

		err := serviceNoBroker.Drop(ctx)

		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
	})

	t.Run("db drop error", func(t *testing.T) {
		expectedErr := assert.AnError
		mockDB.On("Drop", ctx).Return(expectedErr).Once()

		err := service.Drop(ctx)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockDB.AssertExpectations(t)
	})

	t.Run("broker publish error", func(t *testing.T) {
		mockDB.On("Drop", ctx).Return(nil).Once()
		mockBroker.On("Publish", "xkcd.db.dropped", []byte("XKCD DB has been dropped")).Return(assert.AnError).Once()

		err := service.Drop(ctx)

		assert.NoError(t, err) // Drop should succeed even if broker fails
		mockDB.AssertExpectations(t)
		mockBroker.AssertExpectations(t)
	})
}

func TestService_Update(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockDB := &mockDB{}
	mockXKCD := &mockXKCD{}
	mockWords := &mockWords{}
	mockBroker := &mockBroker{}
	ctx := context.Background()

	service, _ := NewService(logger, mockDB, mockXKCD, mockWords, mockBroker, 1)

	t.Run("already running", func(t *testing.T) {
		// Simulate service already running by acquiring the lock
		service.lock.Lock()
		defer service.lock.Unlock()
		defer service.inProgress.Store(false) // cleanup

		err := service.Update(ctx)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrAlreadyExists)
	})

	t.Run("db ids error", func(t *testing.T) {
		mockDB.On("IDs", ctx).Return([]int(nil), assert.AnError).Once()

		err := service.Update(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get existing IDs in DB")
		mockDB.AssertExpectations(t)
	})

	t.Run("xkcd last id error", func(t *testing.T) {
		mockDB.On("IDs", ctx).Return([]int{1, 2}, nil).Once()
		mockXKCD.On("LastID", ctx).Return(0, assert.AnError).Once()

		err := service.Update(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get last ID in XKCD")
		mockDB.AssertExpectations(t)
		mockXKCD.AssertExpectations(t)
	})

	t.Run("successful update with new comics", func(t *testing.T) {
		mockDB.On("IDs", ctx).Return([]int{1}, nil).Once()
		mockXKCD.On("LastID", ctx).Return(3, nil).Once()

		// Mock XKCD Get calls for IDs 2 and 3
		mockXKCD.On("Get", ctx, 2).Return(XKCDInfo{
			ID:          2,
			Description: "test comic 2",
		}, nil).Once()
		mockXKCD.On("Get", ctx, 3).Return(XKCDInfo{
			ID:          3,
			Description: "test comic 3",
		}, nil).Once()

		// Mock Words normalization
		mockWords.On("Norm", ctx, "test comic 2").Return([]string{"test", "comic"}, nil).Once()
		mockWords.On("Norm", ctx, "test comic 3").Return([]string{"test", "another"}, nil).Once()

		// Mock DB Add calls
		mockDB.On("Add", ctx, mock.AnythingOfType("core.Comics")).Return(nil).Twice()

		// Mock broker publish
		mockBroker.On("Publish", "xkcd.db.updated", []byte("XKCD DB has been updated")).Return(nil).Once()

		err := service.Update(ctx)

		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
		mockXKCD.AssertExpectations(t)
		mockWords.AssertExpectations(t)
		mockBroker.AssertExpectations(t)
	})
}

func TestGenerateIDs(t *testing.T) {
	ctx := context.Background()

	t.Run("generate new ids", func(t *testing.T) {
		exists := map[int]bool{1: true, 3: true}
		ch := generateIDs(ctx, 1, 5, exists)

		var result []int
		for id := range ch {
			result = append(result, id)
		}

		assert.Equal(t, []int{2, 4, 5}, result)
	})

	t.Run("no new ids", func(t *testing.T) {
		exists := map[int]bool{1: true, 2: true, 3: true}
		ch := generateIDs(ctx, 1, 3, exists)

		var result []int
		for id := range ch {
			result = append(result, id)
		}

		assert.Empty(t, result)
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		exists := map[int]bool{}
		ch := generateIDs(ctx, 1, 100, exists)

		cancel() // Cancel immediately

		select {
		case <-ch:
			t.Error("channel should be closed on context cancellation")
		default:
			// Channel should be closed
		}
	})
}

// Mock implementations
type mockDB struct {
	mock.Mock
}

func (m *mockDB) Add(ctx context.Context, comics Comics) error {
	args := m.Called(ctx, comics)
	return args.Error(0)
}

func (m *mockDB) Stats(ctx context.Context) (DBStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(DBStats), args.Error(1)
}

func (m *mockDB) Drop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockDB) IDs(ctx context.Context) ([]int, error) {
	args := m.Called(ctx)
	return args.Get(0).([]int), args.Error(1)
}

type mockXKCD struct {
	mock.Mock
}

func (m *mockXKCD) Get(ctx context.Context, id int) (XKCDInfo, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(XKCDInfo), args.Error(1)
}

func (m *mockXKCD) LastID(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

type mockWords struct {
	mock.Mock
}

func (m *mockWords) Norm(ctx context.Context, phrase string) ([]string, error) {
	args := m.Called(ctx, phrase)
	return args.Get(0).([]string), args.Error(1)
}

type mockBroker struct {
	mock.Mock
}

func (m *mockBroker) Publish(topic string, data []byte) error {
	args := m.Called(topic, data)
	return args.Error(0)
}
