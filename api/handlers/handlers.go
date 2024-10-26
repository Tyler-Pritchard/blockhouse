package handlers

import (
	"blockhouse/kafka"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Helper function to validate API key
func validateAPIKey(r *http.Request) bool {
	clientAPIKey := r.Header.Get("X-API-Key")
	expectedAPIKey := os.Getenv("API_KEY")
	if clientAPIKey == expectedAPIKey {
		return true
	}
	log.Printf("Unauthorized access attempt: invalid or missing API key, received: %s", clientAPIKey)
	return false
}

// StartStream creates a new data stream and returns a unique stream_id
func StartStream(w http.ResponseWriter, r *http.Request) {
	// Validate API key
	if !validateAPIKey(r) {
		http.Error(w, "Unauthorized: invalid or missing API key", http.StatusUnauthorized)
		return
	}

	streamID := uuid.New().String()

	response := map[string]string{"stream_id": streamID}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// SendData handles data for an existing stream
func SendData(w http.ResponseWriter, r *http.Request) {
	// Validate API key
	if !validateAPIKey(r) {
		http.Error(w, "Unauthorized: invalid or missing API key", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	streamID := vars["stream_id"]

	clientStreamID := r.Header.Get("X-Stream-ID")
	if clientStreamID != streamID {
		log.Printf("Forbidden: client attempted access to stream %s with stream ID %s", streamID, clientStreamID)
		http.Error(w, "Forbidden: Access to this stream is restricted", http.StatusForbidden)
		return
	}

	// Parse JSON payload
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("Error decoding JSON for stream %s: %v", streamID, err)
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Ensure data is not empty
	if len(data) == 0 {
		log.Printf("No data provided for stream %s", streamID)
		http.Error(w, "No data to process", http.StatusBadRequest)
		return
	}

	// Send data to Kafka
	if err := kafka.SendToKafka(streamID, data); err != nil {
		log.Printf("Kafka error for stream %s: %v", streamID, err)
		http.Error(w, "Failed to send data to Kafka", http.StatusInternalServerError)
		return
	}

	log.Printf("Data sent to Kafka for stream %s", streamID)
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "data accepted"})
}

// GetResults processes and returns results for a stream
func GetResults(w http.ResponseWriter, r *http.Request) {
	// Validate API key
	if !validateAPIKey(r) {
		http.Error(w, "Unauthorized: invalid or missing API key", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	streamID := vars["stream_id"]

	clientStreamID := r.Header.Get("X-Stream-ID")
	if clientStreamID != streamID {
		log.Printf("Forbidden: client attempted access to stream %s with stream ID %s", streamID, clientStreamID)
		http.Error(w, "Forbidden: Access to this stream is restricted", http.StatusForbidden)
		return
	}

	resultChan := make(chan string)
	defer close(resultChan)

	go kafka.ProcessMessages(streamID, resultChan)

	select {
	case result := <-resultChan:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"processed_result": result})
		log.Printf("Processed result sent for stream %s", streamID)
	case <-time.After(5 * time.Second):
		http.Error(w, "No data processed within timeout period", http.StatusGatewayTimeout)
		log.Printf("Timeout on processed data for stream %s", streamID)
	}
}

// WebSocket upgrader setup
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// StreamResults streams processed results over WebSocket
func StreamResults(w http.ResponseWriter, r *http.Request) {
	// Check headers first, then fallback to query parameters if headers are empty
	clientAPIKey := r.Header.Get("X-API-Key")
	if clientAPIKey == "" {
		clientAPIKey = r.URL.Query().Get("X-API-Key")
	}

	streamID := r.Header.Get("X-Stream-ID")
	if streamID == "" {
		streamID = r.URL.Query().Get("X-Stream-ID")
	}

	// Validate API key and stream ID
	expectedAPIKey := os.Getenv("API_KEY")
	if clientAPIKey != expectedAPIKey || streamID == "" {
		log.Printf("Unauthorized WebSocket access attempt: missing or invalid API key, received: %s", clientAPIKey)
		http.Error(w, "Unauthorized: invalid or missing API key", http.StatusUnauthorized)
		return
	}

	// Proceed with WebSocket upgrade and message handling as usual
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	resultChan := make(chan string)
	defer close(resultChan)

	go kafka.ProcessMessages(streamID, resultChan)

	for result := range resultChan {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(result)); err != nil {
			log.Println("Failed to send message over WebSocket:", err)
			break
		}
	}
}
