package db

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"log/slog"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"yadro.com/course/search/core"
)

func TestNew(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	t.Run("connection failure", func(t *testing.T) {
		db, err := New(logger, "invalid:connection:string")
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

func TestDB_Close(t *testing.T) {
	t.Run("close with nil connection", func(t *testing.T) {
		db := &DB{}
		err := db.Close()
		assert.Error(t, err) // Should error because conn is nil
	})
}

func TestDB_Search(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	ctx := context.Background()

	t.Run("successful search with results", func(t *testing.T) {
		// Create sqlmock
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		expectedIDs := []int{1, 3, 5}
		rows := sqlmock.NewRows([]string{"comic_id"}).
			AddRow(1).
			AddRow(3).
			AddRow(5)

		mock.ExpectQuery(`SELECT DISTINCT comic_id FROM comic_words WHERE word = \$1`).
			WithArgs("test").
			WillReturnRows(rows)

		result, err := db.Search(ctx, "test")

		assert.NoError(t, err)
		assert.Equal(t, expectedIDs, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("successful search with no results", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		rows := sqlmock.NewRows([]string{"comic_id"})

		mock.ExpectQuery(`SELECT DISTINCT comic_id FROM comic_words WHERE word = \$1`).
			WithArgs("nonexistent").
			WillReturnRows(rows)

		result, err := db.Search(ctx, "nonexistent")

		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		mock.ExpectQuery(`SELECT DISTINCT comic_id FROM comic_words WHERE word = \$1`).
			WithArgs("test").
			WillReturnError(errors.New("database error"))

		result, err := db.Search(ctx, "test")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_Get(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	ctx := context.Background()

	t.Run("successful get with keywords", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		// Mock comics query
		comicsRows := sqlmock.NewRows([]string{"id", "url"}).
			AddRow(1, "http://example.com/1")

		mock.ExpectQuery(`SELECT id, url FROM comics WHERE id = \$1`).
			WithArgs(1).
			WillReturnRows(comicsRows)

		// Mock keywords query
		keywordsRows := sqlmock.NewRows([]string{"word"}).
			AddRow("hello").
			AddRow("world")

		mock.ExpectQuery(`SELECT word FROM comic_words WHERE comic_id = \$1 ORDER BY word`).
			WithArgs(1).
			WillReturnRows(keywordsRows)

		result, err := db.Get(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "http://example.com/1", result.URL)
		assert.Equal(t, []string{"hello", "world"}, result.Keywords)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("comics not found", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		mock.ExpectQuery(`SELECT id, url FROM comics WHERE id = \$1`).
			WithArgs(999).
			WillReturnError(sql.ErrNoRows)

		result, err := db.Get(ctx, 999)

		assert.Error(t, err)
		assert.Equal(t, core.ErrNotFound, err)
		assert.Equal(t, core.Comics{}, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error on comics query", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		mock.ExpectQuery(`SELECT id, url FROM comics WHERE id = \$1`).
			WithArgs(1).
			WillReturnError(errors.New("database error"))

		result, err := db.Get(ctx, 1)

		assert.Error(t, err)
		assert.Equal(t, core.Comics{}, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error on keywords query", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		// Mock comics query
		comicsRows := sqlmock.NewRows([]string{"id", "url"}).
			AddRow(1, "http://example.com/1")

		mock.ExpectQuery(`SELECT id, url FROM comics WHERE id = \$1`).
			WithArgs(1).
			WillReturnRows(comicsRows)

		// Mock keywords query with error
		mock.ExpectQuery(`SELECT word FROM comic_words WHERE comic_id = \$1 ORDER BY word`).
			WithArgs(1).
			WillReturnError(errors.New("keywords query error"))

		result, err := db.Get(ctx, 1)

		assert.Error(t, err)
		assert.Equal(t, core.Comics{}, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_LastID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	ctx := context.Background()

	t.Run("successful get last ID", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		rows := sqlmock.NewRows([]string{"coalesce"}).
			AddRow(42)

		mock.ExpectQuery(`SELECT coalesce\(max\(id\), 0\) FROM comics`).
			WillReturnRows(rows)

		result, err := db.LastID(ctx)

		assert.NoError(t, err)
		assert.Equal(t, 42, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty database returns 0", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		rows := sqlmock.NewRows([]string{"coalesce"}).
			AddRow(0)

		mock.ExpectQuery(`SELECT coalesce\(max\(id\), 0\) FROM comics`).
			WillReturnRows(rows)

		result, err := db.LastID(ctx)

		assert.NoError(t, err)
		assert.Equal(t, 0, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		mock.ExpectQuery(`SELECT coalesce\(max\(id\), 0\) FROM comics`).
			WillReturnError(errors.New("database error"))

		result, err := db.LastID(ctx)

		assert.Error(t, err)
		assert.Equal(t, 0, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
