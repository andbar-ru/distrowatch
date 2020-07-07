package main

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/andbar-ru/average_color"
	"github.com/andbar-ru/distrowatch"
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
	var query string
	if r.URL.Query().Get("dropout") == "true" {
		query = "SELECT name, SUM(count) AS count, MAX(last_update) AS last_update FROM (SELECT name, count, last_update FROM distrs UNION ALL SELECT name, count, last_update FROM dropout) GROUP BY name"
	} else {
		query = "SELECT name, count, last_update FROM distrs"
	}
	if orderBy := r.URL.Query()["orderBy"]; len(orderBy) > 0 {
		orderByStr, err := getOrderByStr(orderBy)
		if err != nil {
			message := fmt.Sprintf("Invalid query '%s': %s", r.URL.RawQuery, err.Error())
			respondError(w, http.StatusBadRequest, message)
			return
		}
		query += orderByStr
	}
	logger.Debug(query)

	var distrs []Distr
	err := db.Select(&distrs, query)
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

// handleAverageColor handles route /average-colors.
// Sends average color of the last image in config.ImagesDir.
func handleAverageColor(w http.ResponseWriter, r *http.Request) {
	imagesDir := getPath(config.ImagesDir)
	files, err := ioutil.ReadDir(imagesDir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Find the last image.
	var lastImage string
	var lastModTime time.Time
	for _, file := range files {
		ext := strings.ToLower(path.Ext(file.Name()))
		// interested only in images
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".gif" {
			continue
		}
		modTime := file.ModTime()
		if modTime.After(lastModTime) {
			lastModTime = modTime
			lastImage = path.Join(distrowatch.DistrsDir, file.Name())
		}
	}
	if lastImage == "" {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("could not find images in directory %s", imagesDir))
		return
	}
	f, err := os.Open(lastImage)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer closeCheck(f)
	img, _, err := image.Decode(f)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	averageColor := average_color.AverageColor(img)

	var averageColorStr string
	if averageColor.A == 0xff {
		averageColorStr = fmt.Sprintf("#%02x%02x%02x", averageColor.R, averageColor.G, averageColor.B)
	} else {
		averageColorStr = fmt.Sprintf("#%02x%02x%02x%02x", averageColor.R, averageColor.G, averageColor.B, averageColor.A)
	}

	respondJSON(w, http.StatusOK, averageColorStr)
}
