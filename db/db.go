package db

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var ConnectionString string

var db *sql.DB
var insertAccessStatement *sql.Stmt
var upsertAnnounceStatement *sql.Stmt
var announceStatement *sql.Stmt

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
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS announce_fobs (fob TEXT NOT NULL PRIMARY KEY, nickname TEXT NOT NULL, announce BOOL NOT NULL)")
	if err != nil {
		return err
	}

	insertAccessStatement, err = db.Prepare("INSERT INTO access (fob, door, ts, allowed) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	upsertAnnounceStatement, err = db.Prepare("INSERT INTO announce_fobs (fob, nickname, announce) VALUES (?, ?, ?) ON CONFLICT (fob) DO UPDATE SET announce = ?, nickname = ?")
	if err != nil {
		return err
	}
	announceStatement, err = db.Prepare("SELECT announce, nickname FROM announce_fobs WHERE fob = ?")
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

	if _, err := insertAccessStatement.Exec(fob, door, time.UnixMilli(), granted); err != nil {
		return err
	}

	return nil
}

func UpsertAnnounce(fob string, announce bool, nickname string) error {
	if err := prepare(); err != nil {
		return err
	}

	if _, err := upsertAnnounceStatement.Exec(fob, announce, nickname, announce, nickname); err != nil {
		return err
	}

	return nil
}

func IsAnnounceEnabled(fob string) (bool, string, error) {
	if err := prepare(); err != nil {
		return false, "", err
	}

	r := announceStatement.QueryRow(fob)
	var b bool
	var s string
	err := r.Scan(&b, &s)
	if err != nil {
		return false, "", err
	} else {
		return b, s, nil
	}
}
