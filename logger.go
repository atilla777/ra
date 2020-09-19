package main

import (
  "os"
  "log"
  "fmt"
	"github.com/spf13/viper"
)

func logStart() {

  // TODO check is 100 good value?
  logChan = make(chan logMessage, 100)
	go func() {
    logFile := viper.GetString("ra.logs")
    f, err := os.OpenFile(logFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
    defer f.Close()
    if err != nil {
      log.Fatalf("Error open log file: %v", err)
    }
    log.SetOutput(f)
    for l := range logChan {
      log.Println(l)
    }
  }()

  logChan <- logMessage(fmt.Sprint("Ra started"))
}
