package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/update/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {

	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) Close() error {
	if db.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	return db.conn.Close()
}

func (db *DB) Add(ctx context.Context, comics core.Comics) error {
	_, err := db.conn.ExecContext(
		ctx,
		"INSERT INTO comics (id, url, title, transcript, alt, image_url) VALUES($1, $2, $3, $4, $5, $6)",
		comics.ID, comics.ImageURL, comics.Title, comics.Transcript, comics.Alt, comics.ImageURL,
	)
	if err != nil {
		return err
	}

	for _, word := range comics.Words {
		_, err = db.conn.ExecContext(
			ctx,
			"INSERT INTO comic_words (comic_id, word) VALUES($1, $2) ON CONFLICT DO NOTHING",
			comics.ID, word,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) Stats(ctx context.Context) (core.DBStats, error) {
	var stats core.DBStats
	err := db.conn.GetContext(
		ctx, &stats.ComicsFetched,
		"SELECT COUNT(*) FROM comics")
	if err != nil {
		return core.DBStats{}, err
	}
	err = db.conn.GetContext(
		ctx, &stats.WordsTotal,
		"SELECT COUNT(*) FROM comic_words",
	)
	if err != nil {
		return core.DBStats{}, err
	}
	err = db.conn.GetContext(
		ctx, &stats.WordsUnique,
		"SELECT COUNT(DISTINCT word) FROM comic_words",
	)
	if err != nil {
		return core.DBStats{}, err
	}

	return stats, nil
}

func (db *DB) IDs(ctx context.Context) ([]int, error) {
	var IDs []int
	err := db.conn.SelectContext(
		ctx, &IDs,
		"SELECT id FROM comics")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return IDs, nil
}

func (db *DB) Drop(ctx context.Context) error {
	// Delete in correct order due to foreign key constraints
	_, err := db.conn.ExecContext(ctx, "DELETE FROM comic_words")
	if err != nil {
		return err
	}
	_, err = db.conn.ExecContext(ctx, "DELETE FROM comics")
	return err
}
