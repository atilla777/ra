package main

import (
	"flag"
	"github.com/spf13/viper"
	"log"
)

// Load ra configurations from file and set defaults
func loadConfig() {
	confPath := flag.String("conf-path", ".", "config file path")
	confName := flag.String("conf-name", "config", "config file name, may be with ext")
	confExt := flag.String("conf-ext", "yaml", "config file extension")
	verbLog := flag.Bool("vL", false, "verbose mode (write info in log)")
	verbCon := flag.Bool("vC", false, "verbose mode (write info and error in console)")
	flag.Parse()

	// Default ra configs
	viper.SetDefault("ra.verblog", verbLog)              // verbocity in console
	viper.SetDefault("ra.verbcon", verbCon)              // verbocity in log
	viper.SetDefault("ra.workers.queue", 25)             // max queue size (max scan number in queue)
	viper.SetDefault("ra.workers.count", 5)              // max workers count
	viper.SetDefault("ra.workers.responser_attempts", 3) // max attempts to sent job result
	viper.SetDefault("ra.workers.timeout", 12)           // max scan job time before kill nmap
	viper.SetDefault("ra.planer.tick", 5)                // planner start period in sec
	viper.SetDefault("ra.scanner.attempts", 3)           // max attrempts to do done job
	viper.SetDefault("ra.crt", "ra.crt")                 // ra ssl cert
	viper.SetDefault("ra.pem", "ra.key")                 // ra ssl privat key
	viper.SetDefault("ra.sqlite", "sqlite.db")           // sqlite file path
	viper.SetDefault("ra.host", "127.0.0.1")             // ra listen on address
	viper.SetDefault("ra.port", "1323")                  // ra listen on port
	viper.SetDefault("ra.logs", "logs")                  // ra log file
	viper.SetDefault("ra.nmapxml", ".")                  // ran nmap results path
	viper.SetDefault("rism.host", "127.0.0.1")           // rism server address
	viper.SetDefault("rism.port", "3000")                //rism server port
	viper.SetDefault("rism.protocol", "http")            // rism server protocol
	viper.SetDefault("rism.path", "/api/v1/ra_api")      // rism ra API path

	// Viper configs
	viper.AddConfigPath(*confPath)  // path to look for the config file in
	viper.AddConfigPath("/etc/ra/") // path to look for the config file in
	viper.AddConfigPath(".")        // apth to look for config in the working directory
	viper.SetConfigName(*confName)  // name of config file (without extension)
	viper.SetConfigType(*confExt)   // if the config file does not have the extension in the name

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error load config file: %v", err)
	}
}
