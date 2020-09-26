package main

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
)

// Log info and error message to file
func logStart() {
	// TODO check is 100 good value?
	logChan = make(chan raLog, 100)
	go func() {
		logFile := viper.GetString("ra.logs")
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		defer f.Close()
		if err != nil {
			log.Fatalf("Error open log file: %v", err)
		}
		log.SetOutput(f)
		for m := range logChan {
			if m.Lev == "err" || viper.GetBool("ra.verblog") {
				log.Println(m.Mes)
			}
			if viper.GetBool("ra.verbcon") {
				fmt.Println(m.Mes)
			}
		}
	}()

	logChan <- raLog{Mes: fmt.Sprint("Ra started")}
}
