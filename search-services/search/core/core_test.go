package core

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewIndex(t *testing.T) {
	index := NewIndex()

	assert.NotNil(t, index)
	assert.NotNil(t, index.index)
	assert.Equal(t, 0, len(index.index))
}

func TestIndex_Clear(t *testing.T) {
	index := NewIndex()

	index.index["test"] = []int{1, 2, 3}
	index.index["another"] = []int{4, 5}

	index.Clear()

	assert.Equal(t, 0, len(index.index))
	assert.NotNil(t, index.index)
}

func TestIndex_Put(t *testing.T) {
	index := NewIndex()

	index.Put(1, []string{"hello", "world"})
	index.Put(2, []string{"hello", "golang"})

	assert.Equal(t, []int{1, 2}, index.index["hello"])
	assert.Equal(t, []int{1}, index.index["world"])
	assert.Equal(t, []int{2}, index.index["golang"])
}

func TestIndex_Get(t *testing.T) {
	index := NewIndex()

	index.Put(1, []string{"test", "keyword"})

	t.Run("existing keyword", func(t *testing.T) {
		result := index.Get("test")
		assert.Equal(t, []int{1}, result)

		result[0] = 999
		assert.Equal(t, []int{1}, index.index["test"])
	})

	t.Run("non-existing keyword", func(t *testing.T) {
		result := index.Get("nonexistent")
		assert.Equal(t, []int(nil), result)
	})

	t.Run("empty keyword", func(t *testing.T) {
		result := index.Get("")
		assert.Equal(t, []int(nil), result)
	})
}

func TestIndex_Concurrency(t *testing.T) {
	index := NewIndex()

	// Test concurrent access
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			index.Put(i, []string{"concurrent"})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			_ = index.Get("concurrent")
		}
		done <- true
	}()

	<-done
	<-done

	assert.Equal(t, 100, len(index.index["concurrent"]))
}

type mockDB struct {
	mock.Mock
}

func (m *mockDB) Search(ctx context.Context, keyword string) ([]int, error) {
	args := m.Called(ctx, keyword)
	return args.Get(0).([]int), args.Error(1)
}

func (m *mockDB) Get(ctx context.Context, id int) (Comics, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Comics), args.Error(1)
}

func (m *mockDB) LastID(ctx context.Context) (int, error) {
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

func TestNewService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockDB := &mockDB{}
	mockWords := &mockWords{}

	service, err := NewService(logger, mockDB, mockWords)

	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, logger, service.log)
	assert.Equal(t, mockDB, service.db)
	assert.Equal(t, mockWords, service.words)
	assert.NotNil(t, service.index)
}

func TestService_Search(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockDB := &mockDB{}
	mockWords := &mockWords{}
	ctx := context.Background()

	service, _ := NewService(logger, mockDB, mockWords)

	t.Run("successful search", func(t *testing.T) {

		mockWords.On("Norm", ctx, "hello world").Return([]string{"hello", "world"}, nil).Once()
		mockDB.On("Search", ctx, "hello").Return([]int{1, 2}, nil).Once()
		mockDB.On("Search", ctx, "world").Return([]int{1, 3}, nil).Once()

		mockDB.On("Get", ctx, 1).Return(Comics{ID: 1, URL: "url1", Keywords: []string{"hello", "world"}}, nil).Once()
		mockDB.On("Get", ctx, 2).Return(Comics{ID: 2, URL: "url2", Keywords: []string{"hello"}}, nil).Once()
		mockDB.On("Get", ctx, 3).Return(Comics{ID: 3, URL: "url3", Keywords: []string{"world"}}, nil).Once()

		result, err := service.Search(ctx, "hello world", 10)

		assert.NoError(t, err)
		assert.Len(t, result, 3)

		assert.Equal(t, 1, result[0].ID)
		assert.Equal(t, 2, result[0].Score)
		assert.Equal(t, 2, result[1].ID)
		assert.Equal(t, 1, result[1].Score)
		assert.Equal(t, 3, result[2].ID)
		assert.Equal(t, 1, result[2].Score)

		mockWords.AssertExpectations(t)
		mockDB.AssertExpectations(t)
	})

	t.Run("normalization error", func(t *testing.T) {
		mockWords.On("Norm", ctx, "bad phrase").Return([]string(nil), errors.New("normalization error")).Once()

		result, err := service.Search(ctx, "bad phrase", 10)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "normalization error")

		mockWords.AssertExpectations(t)
	})

	t.Run("database search error", func(t *testing.T) {
		mockWords.On("Norm", ctx, "test").Return([]string{"test"}, nil).Once()
		mockDB.On("Search", ctx, "test").Return([]int(nil), errors.New("db error")).Once()

		result, err := service.Search(ctx, "test", 10)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "db error")

		mockWords.AssertExpectations(t)
		mockDB.AssertExpectations(t)
	})
}

func TestService_SearchIndex(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockDB := &mockDB{}
	mockWords := &mockWords{}
	ctx := context.Background()

	service, _ := NewService(logger, mockDB, mockWords)

	// Pre-populate index
	service.index.Put(1, []string{"hello", "world"})
	service.index.Put(2, []string{"hello"})
	service.index.Put(3, []string{"world"})

	t.Run("successful search", func(t *testing.T) {
		mockWords.On("Norm", ctx, "hello world").Return([]string{"hello", "world"}, nil).Once()

		mockDB.On("Get", ctx, 1).Return(Comics{ID: 1, URL: "url1", Keywords: []string{"hello", "world"}}, nil).Once()
		mockDB.On("Get", ctx, 2).Return(Comics{ID: 2, URL: "url2", Keywords: []string{"hello"}}, nil).Once()
		mockDB.On("Get", ctx, 3).Return(Comics{ID: 3, URL: "url3", Keywords: []string{"world"}}, nil).Once()

		result, err := service.SearchIndex(ctx, "hello world", 10)

		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, 1, result[0].ID)
		assert.Equal(t, 2, result[0].Score)
		assert.Equal(t, 2, result[1].ID)
		assert.Equal(t, 1, result[1].Score)
		assert.Equal(t, 3, result[2].ID)
		assert.Equal(t, 1, result[2].Score)

		mockWords.AssertExpectations(t)
		mockDB.AssertExpectations(t)
	})

	t.Run("normalization error", func(t *testing.T) {
		mockWords.On("Norm", ctx, "bad phrase").Return([]string(nil), errors.New("normalization error")).Once()

		result, err := service.SearchIndex(ctx, "bad phrase", 10)

		assert.Error(t, err)
		assert.Nil(t, result)

		mockWords.AssertExpectations(t)
	})
}

func TestService_BuildIndex(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockDB := &mockDB{}
	mockWords := &mockWords{}
	ctx := context.Background()

	service, _ := NewService(logger, mockDB, mockWords)

	t.Run("successful build", func(t *testing.T) {
		mockDB.On("LastID", ctx).Return(3, nil).Once()
		mockDB.On("Get", ctx, 1).Return(Comics{ID: 1, Keywords: []string{"hello", "world"}}, nil).Once()
		mockDB.On("Get", ctx, 2).Return(Comics{ID: 2, Keywords: []string{"golang"}}, nil).Once()
		mockDB.On("Get", ctx, 3).Return(Comics{ID: 3, Keywords: []string{"hello"}}, nil).Once()

		err := service.BuildIndex(ctx)

		assert.NoError(t, err)

		assert.Equal(t, []int{1, 3}, service.index.Get("hello"))
		assert.Equal(t, []int{1}, service.index.Get("world"))
		assert.Equal(t, []int{2}, service.index.Get("golang"))

		mockDB.AssertExpectations(t)
	})

	t.Run("last ID error", func(t *testing.T) {
		mockDB.On("LastID", ctx).Return(0, errors.New("db error")).Once()

		err := service.BuildIndex(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")

		mockDB.AssertExpectations(t)
	})

	t.Run("get comics error", func(t *testing.T) {
		mockDB.On("LastID", ctx).Return(2, nil).Once()
		mockDB.On("Get", ctx, 1).Return(Comics{ID: 1, Keywords: []string{"hello"}}, nil).Once()
		mockDB.On("Get", ctx, 2).Return(Comics{}, errors.New("get error")).Once()

		err := service.BuildIndex(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get error")

		mockDB.AssertExpectations(t)
	})

	t.Run("skip not found comics", func(t *testing.T) {
		mockDB.On("LastID", ctx).Return(2, nil).Once()
		mockDB.On("Get", ctx, 1).Return(Comics{ID: 1, Keywords: []string{"hello"}}, nil).Once()
		mockDB.On("Get", ctx, 2).Return(Comics{}, ErrNotFound).Once()

		err := service.BuildIndex(ctx)

		assert.NoError(t, err)
		assert.Equal(t, []int{1}, service.index.Get("hello"))

		mockDB.AssertExpectations(t)
	})
}

func TestService_ClearIndex(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockDB := &mockDB{}
	mockWords := &mockWords{}

	service, _ := NewService(logger, mockDB, mockWords)

	service.index.Put(1, []string{"test"})

	service.ClearIndex()

	assert.Equal(t, []int(nil), service.index.Get("test"))
}
