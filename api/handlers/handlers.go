package handlers

import (
	"blockhouse/kafka"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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

	clientStreamID := r.Header.Get("X-Stream-ID")
	if clientStreamID != streamID {
		log.Printf("Unauthorized access attempt: client tried to access stream %s with id %s\n", streamID, clientStreamID)
		http.Error(w, "Forbidden: Access to this stream is restricted", http.StatusForbidden)
		return
	}

	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Invalid JSON data for stream %s: %v", streamID, err)
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	err = kafka.SendToKafka(streamID, data)
	if err != nil {
		log.Printf("Failed to send data to Kafka for stream %s: %v", streamID, err)
		http.Error(w, "Failed to send data to Kafka", http.StatusInternalServerError)
		return
	}

	log.Printf("Data successfully sent to Kafka for stream %s", streamID)
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "data accepted"})
}

// GetResults starts the consumer and returns processed results for a stream
func GetResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["stream_id"]

	// Simulate stream ownership check (e.g., could use a token-based check here)
	clientStreamID := r.Header.Get("X-Stream-ID") // Ensure clients pass their unique stream_id in headers
	if clientStreamID != streamID {
		log.Printf("Unauthorized access attempt: client tried to access stream %s with id %s\n", streamID, clientStreamID)
		http.Error(w, "Forbidden: Access to this stream is restricted", http.StatusForbidden)
		return
	}

	// Channel to receive processed messages
	resultChan := make(chan string)
	defer close(resultChan)

	// Start Kafka consumer in a goroutine
	go kafka.ProcessMessages(streamID, resultChan)

	// Set a timeout for receiving a processed message
	select {
	case result := <-resultChan:
		// Send processed result to the client
		response := map[string]string{"processed_result": result}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		log.Printf("Sent processed result for stream %s", streamID)
	case <-time.After(5 * time.Second):
		// Timeout if no message is received in 5 seconds
		http.Error(w, "No data processed within timeout period", http.StatusGatewayTimeout)
		log.Printf("Timeout waiting for processed data from stream %s", streamID)
	}
}

// Set up a WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// StreamResults establishes a WebSocket connection to stream processed Kafka results in real-time
func StreamResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["stream_id"]

	// Validate that the client is using the correct stream_id in the header
	clientStreamID := r.Header.Get("X-Stream-ID")
	if clientStreamID != streamID {
		log.Printf("Unauthorized WebSocket access attempt: client tried to access stream %s with id %s\n", streamID, clientStreamID)
		http.Error(w, "Forbidden: Access to this stream is restricted", http.StatusForbidden)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to set WebSocket upgrade:", err)
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
