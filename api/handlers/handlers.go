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

	// Parse the JSON data
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Send the data to Kafka (using the producer defined in producer.go)
	err = kafka.SendToKafka(streamID, data)
	if err != nil {
		http.Error(w, "Failed to send data to Kafka", http.StatusInternalServerError)
		log.Println("Error producing to Kafka:", err)
		return
	}

	// Respond with a success message
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "data accepted"})
}

// GetResults starts the consumer and returns processed results for a stream
func GetResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["stream_id"]

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

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to set WebSocket upgrade:", err)
		return
	}
	defer conn.Close()

	// Create a channel to receive processed messages
	resultChan := make(chan string)
	defer close(resultChan)

	// Start Kafka consumer in a goroutine to push data to the result channel
	go kafka.ProcessMessages(streamID, resultChan)

	// Listen to the result channel and send data to WebSocket
	for result := range resultChan {
		err := conn.WriteMessage(websocket.TextMessage, []byte(result))
		if err != nil {
			log.Println("Failed to send message over WebSocket:", err)
			break
		}
	}
}
