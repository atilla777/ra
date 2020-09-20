package main

import (
	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"fmt"
	"github.com/spf13/viper"
	"log"
)

const (
	dbSchemaSQL = `CREATE TABLE IF NOT EXISTS jobs (
      id text PRIMARY KEY,
      options text NOT NULL,
      status integer NOT NULL DEFAULT 0,
      attempts integer NOT NULL DEFAULT 0,
      created_at text NOT NULL
    );`
)

func createPool() *sqlitex.Pool {
	sqliteConnStr := fmt.Sprintf(
		"file:%s?cache=shared&mode=rwc&_journal_mode=WAL",
		viper.GetString("ra.sqlite"),
	)
	db, err := sqlitex.Open(sqliteConnStr, 0, 16)
	if err != nil {
		log.Fatalf("Error load config file: %v", err)
	}
	return db
}

func createDatabse() {
	err := execSQL(dbSchemaSQL, nil)
	if err != nil {
		log.Fatalf("Error create/check existence DB: %v", err)
	}
}

func databaseWriter() {
	for wc := range writeChan {
		wc.resultChan <- executeCommand(wc.command, wc.params)
	}
}

func executeCommand(command string, params []string) error {
	// Make slice of interfaces from slice of strings to db.Exec args
	args := make([]interface{}, len(params))
	for i, p := range params {
		args[i] = p
	}
	// Execute SQL command
	err := execSQL(command, nil)
	return err
}

func execSQL(sql string, resultFn func(stmt *sqlite.Stmt) error) error {
	conn := pool.Get(nil)
	defer pool.Put(conn)

	err := sqlitex.Exec(conn, sql, nil)
	if err != nil {
		return err
	}
	return nil
}
