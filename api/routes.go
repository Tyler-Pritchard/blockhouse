package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRoutes initializes the API routes and applies middleware
func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Apply middlewares
	router.Use(LoggingMiddleware)
	router.Use(AuthMiddleware)

	// Define routes
	router.HandleFunc("/stream/start", StartStreamHandler).Methods("POST")
	router.HandleFunc("/stream/{stream_id}/send", SendDataHandler).Methods("POST")
	router.HandleFunc("/stream/{stream_id}/results", GetResultsHandler).Methods("GET")

	// Placeholder route for testing
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the real-time data streaming API!"))
	})

	return router
}
