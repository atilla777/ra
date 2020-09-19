package main

import (
	"database/sql"
	"fmt"
	"github.com/labstack/echo"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"time"
)

const (
	createJobSQL = `INSERT INTO jobs (id, options, status, attempts, created_at)
      VALUES (?, ?, ?, ?, ?)`
	checkExistenceSQL = `SELECT EXISTS (SELECT 1 FROM jobs WHERE id = ?)`
)

// Echo controller: Save recived through API scan job in sqlite database
//func createScan(db *sql.DB) echo.HandlerFunc {
func createScan() echo.HandlerFunc {
	return func(c echo.Context) error {
		// RISM scan background job id from POST form

		db, err := sql.Open("sqlite3", sqliteConnStr)
		defer db.Close()
		if err != nil {
			logChan <- logMessage(fmt.Sprintf("Nmap scan controller open DB error: %s", err))
			return c.String(http.StatusInternalServerError, "Error")
		}
		id := c.FormValue("id")
		exist, err := rowExists(db, id)
		if err != nil {
			logChan <- logMessage(fmt.Sprintf("Nmap scan controller check existence error: %s", err))
			return c.String(http.StatusInternalServerError, "Error")
		}
		if exist {
			return c.String(http.StatusNotAcceptable, "Record already exists.")
		}
		// Scan options from POST form
		options := c.FormValue("options")
		if err := insertRow(id, options); err != nil {
			logChan <- logMessage(fmt.Sprintf("Nmap scan controller save job error: %s", err))
			// TODO add error status
			return c.String(http.StatusInternalServerError, "Error")
		}
		logChan <- logMessage(fmt.Sprintf("Scan %s accepted.", id))
		return c.String(http.StatusOK, "Ok")
	}
}

// Insert job with ID to sqlite database
func insertRow(id string, options string) error {
	resultChan := make(chan error)
	c := writeCommand{
		command:    createJobSQL,
		params:     []string{id, options, "0", "0", time.Now().String()},
		resultChan: resultChan,
	}
	writeChan <- c
	return <-resultChan
}

// Check that job with ID already exists in sqlite database
func rowExists(db *sql.DB, id string) (bool, error) {
	var exist bool
	err := db.QueryRow(checkExistenceSQL, id).Scan(&exist)
	return exist, err
}
