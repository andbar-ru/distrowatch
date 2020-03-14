package main

import (
	"encoding/json"
	"os"
)

// Config stores app configuration.
type Config struct {
	Port         int        `json:"port"`
	DatabasePath string     `json:"databasePath"`
	LogConfig    *LogConfig `json:"log"`
}

// LogConfig stores log configuration.
type LogConfig struct {
	Files []string `json:"files"`
	Level string   `json:"level"`
}

// GetConfig reads config.json and decodes it into Config.
func GetConfig() *Config {
	file, err := os.Open("config.json")
	checkErr(err)
	defer closeCheck(file)

	decoder := json.NewDecoder(file)
	var config Config
	err = decoder.Decode(&config)
	checkErr(err)
	return &config
}
