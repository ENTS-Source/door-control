package db

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var ConnectionString string

var db *sql.DB
var insertStatement *sql.Stmt

func prepare() error {
	if db != nil {
		return nil
	}

	var err error
	db, err = sql.Open("sqlite3", ConnectionString)
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS access (fob TEXT NOT NULL, door TEXT NOT NULL, ts BIGINT NOT NULL, allowed BOOL NOT NULL)")
	if err != nil {
		return err
	}

	insertStatement, err = db.Prepare("INSERT INTO access (fob, door, ts, allowed) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}

	return nil
}

func Stop() {
	_ = db.Close()
}

func InsertAccess(door string, fob string, time time.Time, granted bool) error {
	if err := prepare(); err != nil {
		return err
	}

	if _, err := insertStatement.Exec(fob, door, time.UnixMilli(), granted); err != nil {
		return err
	}

	return nil
}
