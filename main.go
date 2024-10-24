package main

import (
	"blockhouse/api"
	"blockhouse/config"
	"log"
	"net/http"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Set up routes
	router := api.SetupRoutes()

	// Log that the server is starting
	log.Println("Server is starting on port 8080...")

	// Start the server
	log.Fatal(http.ListenAndServe(":8080", router))
}
