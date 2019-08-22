package show

import (
	"database/sql"
	"os"
	"path"

	// Register sqlite3.
	_ "github.com/mattn/go-sqlite3"
)

func getDB() (*sql.DB, error) {
	database := os.Getenv("DISTRS_DATABASE")
	if database == "" {
		database = path.Join(os.Getenv("HOME"), "Images/distrs/db.sqlite3")
	}
	// Check if database exists.
	if _, err := os.Stat(database); os.IsNotExist(err) {
		return nil, err
	}
	// Open database.
	db, err := sql.Open("sqlite3", database)
	if err != nil {
		return nil, err
	}
	return db, nil
}
