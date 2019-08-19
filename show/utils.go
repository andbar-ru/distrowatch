package show

import (
	"database/sql"
	"os"

	// Register sqlite3.
	_ "github.com/mattn/go-sqlite3"
)

func getDB(database string) (*sql.DB, error) {
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
