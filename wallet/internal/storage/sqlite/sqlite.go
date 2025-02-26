package sqlite

import (
	"database/sql"
	"fmt"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	Db *sql.DB
}

func New(path string) (*Storage, error) {
	const fn = "storage.sqlite.New"

	//open bd
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	//убрать
	if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	//check bd with ping
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS wallet(
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			balance FLOAT DEFAULT 0.0,
			status TEXT NOT NULL CHECK (status IN ('active', 'inactive')));
		CREATE INDEX IF NOT EXISTS idx_name ON wallet(name);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	return &Storage{Db: db}, nil
}
