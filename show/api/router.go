package main

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// RequestLogger wraps handler and logs request and time it takes.
func RequestLogger(handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler.ServeHTTP(w, r)
		logger.Info("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

// NewRouter returns router.
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		handler := RequestLogger(route.HandlerFunc)
		router.
			Name(route.Name).
			Methods(route.Method).
			Path(route.Pattern).
			Handler(handler)
	}
	return router
}

// Route represents a single route.
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes represents set of routes.
type Routes []Route

var routes = Routes{
	Route{"Status", "GET", "/status", handleStatus},
	Route{"Distrs", "GET", "/distrs", handleDistrs},
}
