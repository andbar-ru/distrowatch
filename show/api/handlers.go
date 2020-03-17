package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// respondJSON makes response with payload in json format.
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	response, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			logger.Panic(err)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write([]byte(response))
	if err != nil {
		logger.Panic(err)
	}
}

// respondError makes error response with payload in json format.
func respondError(w http.ResponseWriter, statusCode int, message string) {
	if statusCode >= 500 {
		logger.Error("%d: %s", statusCode, message)
	} else {
		logger.Warning("%d: %s", statusCode, message)
	}
	respondJSON(w, statusCode, map[string]string{"error": message})
}

// handleStatus handles route /status.
func handleStatus(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{"version": apiVersion}
	respondJSON(w, http.StatusOK, data)
}

// handleDistrs handles route /distrs.
func handleDistrs(w http.ResponseWriter, r *http.Request) {
	query, err := buildSQLQuery("distrs", r.URL.Query(), map[string]bool{"orderBy": true})
	if err != nil {
		message := fmt.Sprintf("Invalid query '%s': %s", r.URL.RawQuery, err.Error())
		respondError(w, http.StatusBadRequest, message)
		return
	}
	var distrs []Distr
	logger.Debug(query)
	err = db.Select(&distrs, query)
	if err != nil {
		if strings.Contains(err.Error(), "no such column") {
			respondError(w, http.StatusBadRequest, err.Error())
		} else {
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondJSON(w, http.StatusOK, distrs)
}

// handleCoords handles route /coords.
func handleCoords(w http.ResponseWriter, r *http.Request) {
	query, err := buildSQLQuery("coords", r.URL.Query(), allDBQueryParams)
	if err != nil {
		message := fmt.Sprintf("Invalid query '%s': %s", r.URL.RawQuery, err.Error())
		respondError(w, http.StatusBadRequest, message)
		return
	}
	var coords []map[string]interface{}
	logger.Debug(query)
	rows, err := db.Queryx(query)
	if err != nil {
		if strings.Contains(err.Error(), "no such column") {
			respondError(w, http.StatusBadRequest, err.Error())
		} else {
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	for rows.Next() {
		coord := make(map[string]interface{})
		err := rows.MapScan(coord)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		coords = append(coords, coord)
	}
	respondJSON(w, http.StatusOK, coords)
}
