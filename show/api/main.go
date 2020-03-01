package main

import (
	"fmt"
	"net/http"
)

var (
	logger *Logger
	config *Config
)

func main() {
	config = GetConfig()
	logger = NewLogger(config.LogConfig)
	router := NewRouter()
	hostPort := fmt.Sprintf("localhost:%d", config.Port)
	logger.Printf("Starting service on %s\n", hostPort)
	logger.Fatal(http.ListenAndServe(hostPort, router))
}
