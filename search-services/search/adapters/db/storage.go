package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/search/core"
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

func (db *DB) Search(ctx context.Context, keyword string) ([]int, error) {
	var IDs []int
	err := db.conn.SelectContext(
		ctx, &IDs,
		"SELECT DISTINCT comic_id FROM comic_words WHERE word = $1",
		keyword,
	)

	return IDs, err
}

type Comics struct {
	ID  int    `db:"id"`
	URL string `db:"url"`
}

func (db *DB) Get(ctx context.Context, id int) (core.Comics, error) {
	var comics Comics
	err := db.conn.GetContext(
		ctx, &comics,
		"SELECT id, url FROM comics WHERE id = $1",
		id,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return core.Comics{}, core.ErrNotFound
	}
	if err != nil {
		return core.Comics{}, err
	}
	var keywords []string
	err = db.conn.SelectContext(
		ctx, &keywords,
		"SELECT word FROM comic_words WHERE comic_id = $1 ORDER BY word",
		id,
	)
	if err != nil {
		return core.Comics{}, err
	}

	return core.Comics{ID: comics.ID, URL: comics.URL, Keywords: keywords}, nil
}

func (db *DB) LastID(ctx context.Context) (int, error) {
	var ID int
	err := db.conn.GetContext(
		ctx, &ID,
		"SELECT coalesce(max(id), 0) FROM comics",
	)

	return ID, err
}
