package db

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"yadro.com/course/update/core"
)

func TestNew_DBConnection(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("invalid address", func(t *testing.T) {
		db, err := New(logger, "invalid://address")

		// Should fail to connect
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

func TestDB_Add(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	ctx := context.Background()

	t.Run("successful add", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		comics := core.Comics{
			ID:         1,
			ImageURL:   "http://example.com/1.png",
			Title:      "Test Comic",
			Transcript: "Test transcript",
			Alt:        "Test alt text",
			Words:      []string{"hello", "world"},
		}

		// Mock comics insert
		mock.ExpectExec(`INSERT INTO comics \(id, url, title, transcript, alt, image_url\) VALUES\(\$1, \$2, \$3, \$4, \$5, \$6\)`).
			WithArgs(1, "http://example.com/1.png", "Test Comic", "Test transcript", "Test alt text", "http://example.com/1.png").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock words inserts
		mock.ExpectExec(`INSERT INTO comic_words \(comic_id, word\) VALUES\(\$1, \$2\) ON CONFLICT DO NOTHING`).
			WithArgs(1, "hello").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(`INSERT INTO comic_words \(comic_id, word\) VALUES\(\$1, \$2\) ON CONFLICT DO NOTHING`).
			WithArgs(1, "world").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = db.Add(ctx, comics)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("comics insert error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		comics := core.Comics{
			ID:    1,
			Words: []string{"hello"},
		}

		mock.ExpectExec(`INSERT INTO comics .*`).
			WithArgs(1, "", "", "", "", "").
			WillReturnError(errors.New("insert error"))

		err = db.Add(ctx, comics)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("words insert error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		comics := core.Comics{
			ID:    1,
			Words: []string{"hello", "world"},
		}

		// Mock successful comics insert
		mock.ExpectExec(`INSERT INTO comics .*`).
			WithArgs(1, "", "", "", "", "").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock failed words insert
		mock.ExpectExec(`INSERT INTO comic_words .*`).
			WithArgs(1, "hello").
			WillReturnError(errors.New("words insert error"))

		err = db.Add(ctx, comics)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_Stats(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	ctx := context.Background()

	t.Run("successful stats", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		// Mock comics count
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comics`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))

		// Mock total words count
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comic_words`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(500))

		// Mock unique words count
		mock.ExpectQuery(`SELECT COUNT\(DISTINCT word\) FROM comic_words`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(250))

		stats, err := db.Stats(ctx)

		assert.NoError(t, err)
		assert.Equal(t, 100, stats.ComicsFetched)
		assert.Equal(t, 500, stats.WordsTotal)
		assert.Equal(t, 250, stats.WordsUnique)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("comics count error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comics`).
			WillReturnError(errors.New("comics count error"))

		stats, err := db.Stats(ctx)

		assert.Error(t, err)
		assert.Equal(t, core.DBStats{}, stats)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_IDs(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	ctx := context.Background()

	t.Run("successful get IDs", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		expectedIDs := []int{1, 2, 3, 5}
		rows := sqlmock.NewRows([]string{"id"}).
			AddRow(1).
			AddRow(2).
			AddRow(3).
			AddRow(5)

		mock.ExpectQuery(`SELECT id FROM comics`).
			WillReturnRows(rows)

		result, err := db.IDs(ctx)

		assert.NoError(t, err)
		assert.Equal(t, expectedIDs, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty database", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		rows := sqlmock.NewRows([]string{"id"})

		mock.ExpectQuery(`SELECT id FROM comics`).
			WillReturnRows(rows)

		result, err := db.IDs(ctx)

		assert.NoError(t, err)
		assert.Nil(t, result)
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

		mock.ExpectQuery(`SELECT id FROM comics`).
			WillReturnError(errors.New("database error"))

		result, err := db.IDs(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_Drop(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	ctx := context.Background()

	t.Run("successful drop", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		// Mock comic_words delete
		mock.ExpectExec(`DELETE FROM comic_words`).
			WillReturnResult(sqlmock.NewResult(0, 50))

		// Mock comics delete
		mock.ExpectExec(`DELETE FROM comics`).
			WillReturnResult(sqlmock.NewResult(0, 10))

		err = db.Drop(ctx)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("comic_words delete error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		mock.ExpectExec(`DELETE FROM comic_words`).
			WillReturnError(errors.New("delete error"))

		err = db.Drop(ctx)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("comics delete error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

		db := &DB{
			log:  logger,
			conn: sqlxDB,
		}

		mock.ExpectExec(`DELETE FROM comic_words`).
			WillReturnResult(sqlmock.NewResult(0, 50))

		mock.ExpectExec(`DELETE FROM comics`).
			WillReturnError(errors.New("delete error"))

		err = db.Drop(ctx)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
