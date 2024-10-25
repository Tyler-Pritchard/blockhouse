package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Handler to start a new stream
func StartStream(w http.ResponseWriter, r *http.Request) {
	// Placeholder response
	response := map[string]string{"message": "Stream started"}
	json.NewEncoder(w).Encode(response)
	log.Println("StartStream endpoint hit")
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
