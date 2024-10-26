package api

import (
	"blockhouse/api/handlers"
	"blockhouse/api/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRoutes initializes the API routes and applies middleware
func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Apply middlewares
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.AuthMiddleware)

	// Define routes
	router.HandleFunc("/stream/start", handlers.StartStream).Methods("POST")
	router.HandleFunc("/stream/{stream_id}/send", handlers.SendData).Methods("POST")
	router.HandleFunc("/stream/{stream_id}/results", handlers.GetResults).Methods("GET")
	router.HandleFunc("/ws/{stream_id}", handlers.StreamResults).Methods("GET")

	// Placeholder route for testing
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the real-time data streaming API!"))
	})

	return router
}
