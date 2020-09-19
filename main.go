// This is scan agent for RISM server

// Command to test api:
// curl -v -H "Authorization: Bearer secret" -X POST http://localhost:1323/scans -d "id=1" -d "options=127.0.0.1"

package main

import (
	"database/sql"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"log"
	"sync"
	"time"
)

type Job struct {
	Id        string
	Options   string
	Attempmts int
}

type logMessage string

var logChan chan logMessage
var jobChan chan Job
var mutex = &sync.Mutex{}
var sqliteConnStr string
var debugMode bool = true

type writeCommand struct {
	command    string
	params     []string
	resultChan chan error
}

var writeChan chan writeCommand

func main() {
	// Make ra initial configuration
	if err := loadConfig(); err != nil {
		log.Fatalf("Error load config file: %v", err)
	}
	sqliteConnStr = fmt.Sprintf(
		"file:%s?cache=shared&mode=rwc&_journal_mode=WAL",
		viper.GetString("ra.sqlite"),
	)

	// Create database if not exist
	if err := createDatabse(); err != nil {
		log.Fatalf("Error create/check existence DB: %v", err)
	}

	// Start logging in file
	logStart()

	// Start database writer
	writeChan = make(chan writeCommand, 100)
	go databaseWriter()

	// Initialize background scans workers pool
	jobChan = make(chan Job, viper.GetInt("ra.workers.queue"))
	for i := 0; i < viper.GetInt("ra.workers.count"); i++ {
		// Variable i can be used to track what goroutine make scan job
		go worker(i)
	}

	// Initilize background job planner (it will create queue in channel and send result to RISM server)
	ticker := time.NewTicker(time.Second * viper.GetDuration("ra.workers.tick"))
	go func() {
		db, err := sql.Open("sqlite3", sqliteConnStr)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		for _ = range ticker.C {
			if err := sendJobsToQueue(db); err != nil {
				logChan <- logMessage(fmt.Sprintf("Planner error: %s", err))
			}
			if err := sendResults(db); err != nil {
				logChan <- logMessage(fmt.Sprintf("Responser error: %s", err))
			}
		}
	}()

	// Initialize Echo web framework
	e := echo.New()
	e.POST("/scans", createScan())
	e.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == viper.GetString("ra.secret"), nil
	}))
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", viper.GetString("ra.host"), viper.GetString("ra.port"))))
}
