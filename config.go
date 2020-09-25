package main

import (
	"flag"
	"github.com/spf13/viper"
	"log"
)

func loadConfig() {
	confPath := flag.String("conf-path", "s", "config file path")
	confName := flag.String("conf-name", "config", "config file name")
	confExt := flag.String("conf-ext", "yaml", "config file extension")
	verbLog := flag.Bool("vL", false, "verbose mode (in log)")
	verbCon := flag.Bool("vC", false, "verbose mode (in console)")
	flag.Parse()

	// Default ra configs
	viper.SetDefault("ra.verblog", verbLog)              // verbocity in console
	viper.SetDefault("ra.verbcon", verbCon)              // verbocity in log
	viper.SetDefault("ra.workers.queue", 25)             // max queue size
	viper.SetDefault("ra.workers.count", 5)              // max workers count
	viper.SetDefault("ra.workers.scanner_attempts", 3)   // max attrempts to do done job
	viper.SetDefault("ra.workers.responser_attempts", 3) // max attrempts to sent job result
	viper.SetDefault("ra.workers.tick", 5)               // planner start period in sec
	viper.SetDefault("ra.crt", "ra.crt")
	viper.SetDefault("ra.pem", "ra.key")
	viper.SetDefault("ra.host", "127.0.0.1")
	viper.SetDefault("ra.sqlite", "sqlite.db") // sqlite file path
	viper.SetDefault("ra.port", "1323")
	viper.SetDefault("ra.logs", "logs")
	viper.SetDefault("ra.nmapxml", ".")
	viper.SetDefault("rism.host", "127.0.0.1")
	viper.SetDefault("rism.port", "3000")
	viper.SetDefault("rism.protocol", "http")
	viper.SetDefault("rism.path", "/api/v1/ra_api")

	// Viper configs
	viper.AddConfigPath(*confPath)  // path to look for the config file in
	viper.AddConfigPath("/etc/ra/") // path to look for the config file in
	viper.AddConfigPath(".")        // optionally look for config in the working directory
	viper.SetConfigName(*confName)  // name of config file (without extension)
	viper.SetConfigType(*confExt)   // REQUIRED if the config file does not have the extension in the name

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error load config file: %v", err)
	}
}

func vL() bool {
	return viper.GetBool("ra.verblog")
}

func vC() bool {
	return viper.GetBool("ra.verblog")
}
