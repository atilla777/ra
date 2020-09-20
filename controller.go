package main

import (
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"time"
)

const (
	checkExistenceSQL = `SELECT EXISTS (SELECT 1 FROM jobs WHERE id = ?)`
)

// Echo controller: Save recived through API scan job in sqlite database
//func createScan(db *sql.DB) echo.HandlerFunc {
func createScan() echo.HandlerFunc {
	return func(c echo.Context) error {
		// RISM scan background job id from POST form
		id := c.FormValue("id")
		exist, err := cehckIdExistence(id)
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
	createJobSQL := `INSERT INTO jobs (id, options, status, attempts, created_at)
      VALUES (?, ?, ?, ?, ?)`
	err := execSQL(createJobSQL, nil, id, options, 0, 0, time.Now().String())
	if err != nil {
		return err
	}
	return nil
}
