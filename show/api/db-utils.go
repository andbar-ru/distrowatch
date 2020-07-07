package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	// Register sqlite3.
	_ "github.com/mattn/go-sqlite3"
)

var (
	allDBQueryParams = map[string]bool{
		"columns": true,
		"orderBy": true,
		"limit":   true,
	}
	columnRgx = regexp.MustCompile(`^\w+$`)
	limitRgx  = regexp.MustCompile(`^\d+$`)
)

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

// getOrderByStr converts request.URL.Query()["orderBy"] to string like "ORDER BY column1 ASC, column2 DESC".
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

// buildSQLQuery composes SQL query using request query parameters.
func buildSQLQuery(table string, query url.Values, allowedParams map[string]bool) (string, error) {
	sqlQuery := "SELECT "

	if columns := query.Get("columns"); allowedParams["columns"] && columns != "" {
		columnsStr, err := getColumnsStr(columns)
		if err != nil {
			return "", err
		}
		sqlQuery += columnsStr
	} else {
		sqlQuery += "*"
	}
	sqlQuery += " FROM " + table

	if orderBy := query["orderBy"]; allowedParams["orderBy"] && len(orderBy) > 0 {
		orderByStr, err := getOrderByStr(orderBy)
		if err != nil {
			return "", err
		}
		sqlQuery += orderByStr
	}

	if limit := query.Get("limit"); allowedParams["limit"] && limit != "" {
		if !limitRgx.MatchString(limit) {
			return "", fmt.Errorf("limit must be number, got '%s'", limit)
		}
		sqlQuery += " LIMIT " + limit
	}

	return sqlQuery, nil
}
