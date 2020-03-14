package main

import (
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	// Register sqlite3.
	_ "github.com/mattn/go-sqlite3"
)

func checkErr(err error) {
	if err != nil {
		if logger == nil {
			log.Panic(err)
		} else {
			logger.Panic(err)
		}
	}
}

func closeCheck(c io.Closer) {
	err := c.Close()
	checkErr(err)
}

func getPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	} else if strings.HasPrefix(path, "~") {
		currentUser, err := user.Current()
		checkErr(err)
		return strings.Replace(path, "~", currentUser.HomeDir, 1)
	} else {
		execPath, err := os.Executable()
		checkErr(err)
		return filepath.Join(filepath.Dir(execPath), path)
	}
}

// getDB opens and returns sqlite database specified in config.
// Consumers have to close the database.
func getDB() (*sqlx.DB, error) {
	databasePath := getPath(config.DatabasePath)
	// Check if database exists.
	_, err := os.Stat(databasePath)
	if err != nil {
		return nil, err
	}
	// Open database.
	db, err := sqlx.Connect("sqlite3", databasePath)
	if err != nil {
		return nil, err
	}
	return db, nil
}
