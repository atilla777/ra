// This is scan agent for RISM server

//Warnings:
// 1. Timeout on http session in responser not set, hence it may cause some problems with dead sessions between ra and rism
// 2. Huge count scan results sent by ra to rism can make dos attack on rism

// Command to test api:
// curl -v -H "Authorization: Bearer secret" -X POST http://localhost:1323/scans -d "id=1" -d "options=127.0.0.1"

package main

import (
	"crawshaw.io/sqlite/sqlitex"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
	"sync"
)

type Job struct {
	Id       string
	Options  string
	Attempts int
}

type raLog struct {
	Lev string
	Mes string
}

var logChan chan raLog
var jobChan chan Job
var mutex = &sync.Mutex{}
var debugMode bool = true

type writeCommand struct {
	command    string
	params     []string
	resultChan chan error
}

var writeChan chan writeCommand

var pool *sqlitex.Pool

func main() {
	// Make ra initial configuration
	loadConfig()

	// Create database connections pool
	pool = createPool()

	// Create database if not exist
	createDatabse()

	// Start logging in file
	logStart()

	// Initialize background scans workers pool
	jobChan = make(chan Job, viper.GetInt("ra.workers.queue"))
	for i := 0; i < viper.GetInt("ra.workers.count"); i++ {
		// Variable i can be used to track what goroutine make scan job
		go worker(i)
	}

	// Initilize background job planner (it will create queue in channel and send result to RISM server)
	go startPlanner()

	// Initialize Echo web framework
	e := echo.New()
	e.POST("/scans", createScan())
	e.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == viper.GetString("ra.secret"), nil
	}))
	if viper.GetString("ra.protocol") == "http" {
		e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", viper.GetString("ra.host"), viper.GetString("ra.port"))))
	} else {
		c := viper.GetString("ra.crt")
		k := viper.GetString("ra.key")
		e.Logger.Fatal(e.StartTLS(fmt.Sprintf("%s:%s", viper.GetString("ra.host"), viper.GetString("ra.port")), c, k))
	}
}
