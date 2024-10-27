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

	// Apply global middlewares
	router.Use(
		middleware.LoggingMiddleware,
		middleware.AuthMiddleware,
		middleware.APIKeyAuthMiddleware,
	)

	// Define API routes with HTTP methods
	apiRoutes := router.PathPrefix("/stream").Subrouter()
	apiRoutes.HandleFunc("/start", handlers.StartStream).Methods(http.MethodPost)
	apiRoutes.HandleFunc("/{stream_id}/send", handlers.SendData).Methods(http.MethodPost)
	apiRoutes.HandleFunc("/{stream_id}/results", handlers.GetResults).Methods(http.MethodGet)

	// Define WebSocket route
	router.HandleFunc("/ws/{stream_id}", handlers.StreamResults).Methods(http.MethodGet)

	// Root route for API description
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the real-time data streaming API!"))
	}).Methods(http.MethodGet)

	return router
}
