package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func databaseWriter() {
	// TODO move param to viper
	db, err := sql.Open("sqlite3", sqliteConnStr)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	for wc := range writeChan {
		wc.resultChan <- executeCommand(db, wc.command, wc.params)
	}
}

func executeCommand(db *sql.DB, command string, params []string) error {
	args := make([]interface{}, len(params))
	for i, p := range params {
		args[i] = p
	}
	_, err := db.Exec(command, args...)
	return err
}
