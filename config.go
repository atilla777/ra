package main

import (
	"github.com/spf13/viper"
)

func loadConfig() error {
	// Default ra configs
	viper.SetDefault("ra.workers.queue", 25) // max queue size
	viper.SetDefault("ra.workers.count", 5) // max workers count
	viper.SetDefault("ra.workers.attempts", 3) // max attrempts to do done job
	viper.SetDefault("ra.workers.tick", 5) // planner start period in sec
	viper.SetDefault("ra.host", "127.0.0.1")
	viper.SetDefault("ra.sqlite", "sqlite.db") // sqlite file path
	viper.SetDefault("ra.port", "1323")
	viper.SetDefault("ra.logs", "logs")
	viper.SetDefault("rism.host", "127.0.0.1")
	viper.SetDefault("rism.port", "3000")
	viper.SetDefault("rism.protocol", "http")
	viper.SetDefault("rism.path", "/api/v1/ra")
  
  // Viper configs
	viper.AddConfigPath("/etc/ra/") // path to look for the config file in
	viper.AddConfigPath(".")        // optionally look for config in the working directory
	viper.SetConfigName("config")   // name of config file (without extension)
	viper.SetConfigType("yaml")     // REQUIRED if the config file does not have the extension in the name

	return viper.ReadInConfig()
}
