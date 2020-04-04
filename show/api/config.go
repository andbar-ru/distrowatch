package main

import (
	"encoding/json"
	"os"
)

// Config stores app configuration.
type Config struct {
	ListenAddress string     `json:"listenAddress"`
	DatabasePath  string     `json:"databasePath"`
	LogConfig     *LogConfig `json:"log"`
	ImagesDir     string     `json:"imagesDir"`
}

// LogConfig stores log configuration.
type LogConfig struct {
	Files []string `json:"files"`
	Level string   `json:"level"`
}

// GetConfig reads config.json and decodes it into Config.
func GetConfig() *Config {
	configPath := os.Getenv("DISTROWATCH_CONFIG")
	if configPath == "" {
		configPath = "config.json"
	}
	file, err := os.Open(configPath)
	checkErr(err)
	defer closeCheck(file)

	decoder := json.NewDecoder(file)
	var config Config
	err = decoder.Decode(&config)
	checkErr(err)
	return &config
}
