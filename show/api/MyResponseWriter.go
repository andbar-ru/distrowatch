package main

import "net/http"

// MyResponseWriter stores status code.
type MyResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader overrides ResponseWriter.WriteHeader to store status code.
func (w *MyResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write overrides ResponseWriter.Write: calls WriteHeader if it didn't called.
func (w *MyResponseWriter) Write(data []byte) (int, error) {
	if w.statusCode == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}

// StatusCode returns stored status code.
func (w *MyResponseWriter) StatusCode() int {
	return w.statusCode
}
