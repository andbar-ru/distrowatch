package main

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/jmoiron/sqlx"
)

const (
	apiVersion = "1.0"
)

var (
	logger *Logger
	config *Config
	db     *sqlx.DB
)

func main() {
	config = GetConfig()
	logger = NewLogger(config.LogConfig)
	var err error
	db, err = getDB()
	checkErr(err)
	defer closeCheck(db)
	router := NewRouter()
	listenAddress := config.ListenAddress
	logger.Printf("Starting service on %s\n", listenAddress)

	logger.Fatal(http.ListenAndServe(listenAddress, handlers.CORS()(router)))
}
