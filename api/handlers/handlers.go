package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// StartStream creates a new data stream and returns a unique stream_id
func StartStream(w http.ResponseWriter, r *http.Request) {
	// Generate a unique stream ID
	streamID := uuid.New().String()

	// Prepare the response
	response := map[string]string{
		"stream_id": streamID,
	}

	// Encode the response as JSON and send it
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Handler to send data to an existing stream
func SendData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["stream_id"]

	// Placeholder response
	response := map[string]string{"message": "Data sent to stream", "stream_id": streamID}
	json.NewEncoder(w).Encode(response)
	log.Println("SendData endpoint hit for stream:", streamID)
}

// Handler to get results of a stream
func GetResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["stream_id"]

	// Placeholder response
	response := map[string]string{"message": "Results retrieved for stream", "stream_id": streamID}
	json.NewEncoder(w).Encode(response)
	log.Println("GetResults endpoint hit for stream:", streamID)
}
