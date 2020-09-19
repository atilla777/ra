package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

const (
	dbSchemaSQL = `CREATE TABLE IF NOT EXISTS jobs (
      id text PRIMARY KEY,
      options text NOT NULL,
      status integer NOT NULL DEFAULT 0,
      attempts integer NOT NULL DEFAULT 0,
      created_at text NOT NULL
    )`
)

func createDatabse() error {
	db, err := sql.Open("sqlite3", sqliteConnStr)
  if err != nil {
    return err
  }
  if _, err := db.Exec(dbSchemaSQL); err != nil {
		return err
	}
  defer db.Close()
  return nil
}
