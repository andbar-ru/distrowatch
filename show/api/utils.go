package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	// Register sqlite3.
	_ "github.com/mattn/go-sqlite3"
)

var (
	columnRgx = regexp.MustCompile(`^\w+$`)
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

// getOrderStr converts request.URL.Query()["orderBy"] to string like "ORDER BY column1 ASC, column2 DESC".
func getOrderByStr(params []string) (string, error) {
	orderByStr := " ORDER BY"

	for i, param := range params {
		columns := strings.Split(param, ",")
		for j, column := range columns {
			parts := strings.Split(column, " ")
			col := parts[0]
			if !columnRgx.MatchString(col) {
				return "", fmt.Errorf("Invalid column name '%s'", col)
			}
			if i == 0 && j == 0 {
				orderByStr += " `" + col + "`"
			} else {
				orderByStr += ", `" + col + "`"
			}
			if len(parts) > 1 {
				order := strings.ToUpper(parts[1])
				if order != "ASC" && order != "DESC" {
					return "", fmt.Errorf("Sorting order must be 'ASC' or 'DESC', got '%s'", parts[1])
				}
				orderByStr += " " + order
			}
		}
	}
	return orderByStr, nil
}

// getColumnsStr converts request.URL.Query().Get("columns") to string like "column1, column2".
func getColumnsStr(columns string) (string, error) {
	var columnsStr string
	cols := strings.Split(columns, ",")
	for i, col := range cols {
		if !columnRgx.MatchString(col) {
			return "", fmt.Errorf("Invalid column name '%s'", col)
		}
		if i == 0 {
			columnsStr += "`" + col + "`"
		} else {
			columnsStr += ", `" + col + "`"
		}
	}
	return columnsStr, nil
}
