package main

import (
	"encoding/json"
	"net/http"
)

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
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
	w.WriteHeader(status)
	_, err = w.Write([]byte(response))
	if err != nil {
		logger.Panic(err)
	}
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{"version": "1.0"}
	respondJSON(w, http.StatusOK, data)
}
