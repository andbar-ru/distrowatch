package distrowatch

import (
	"database/sql"
	"os"
	"path"

	// Register sqlite3.
	_ "github.com/mattn/go-sqlite3"
)

var (
	// DistrsDir is directory where distrs images and sqlite database store.
	DistrsDir = path.Join(os.Getenv("HOME"), "Images/distrs")
)

// GetDB opens and returns sqlite database from the predefined place.
// Consumers have to close the database.
func GetDB() (*sql.DB, error) {
	database := os.Getenv("DISTRS_DATABASE")
	if database == "" {
		database = path.Join(DistrsDir, "db.sqlite3")
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
