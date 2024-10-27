package handlers

import (
	"blockhouse/kafka"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	once         sync.Once
	cachedAPIKey string
)

// Cache API key from the environment once at startup
func loadAPIKey() {
	cachedAPIKey = os.Getenv("API_KEY")
	if cachedAPIKey == "" {
		log.Println("Warning: API_KEY is not set in the environment.")
	}
}

// Validate API key against cached value
func ValidateAPIKey(r *http.Request) bool {
	once.Do(loadAPIKey)

	clientAPIKey := r.Header.Get("X-API-Key")
	return clientAPIKey == cachedAPIKey
}

// Structure for response when creating a new stream
type StreamResponse struct {
	StreamID string `json:"stream_id"`
}

// StartStream creates a new data stream and returns a unique stream_id
func StartStream(w http.ResponseWriter, r *http.Request) {
	// Validate API key
	if !ValidateAPIKey(r) {
		http.Error(w, "Unauthorized: invalid or missing API key", http.StatusUnauthorized)
		return
	}

	// Generate unique stream ID and respond
	streamID := uuid.New().String()
	response := StreamResponse{StreamID: streamID}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Reuse json encoder
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Structure for response when sending data to a stream
type SendDataResponse struct {
	Status string `json:"status"`
}

// SendData sends data to an existing Kafka stream
func SendData(w http.ResponseWriter, r *http.Request) {
	if !ValidateAPIKey(r) {
		http.Error(w, "Unauthorized: invalid or missing API key", http.StatusUnauthorized)
		return
	}

	// Extract stream ID from URL
	vars := mux.Vars(r)
	streamID := vars["stream_id"]
	if clientStreamID := r.Header.Get("X-Stream-ID"); clientStreamID != streamID {
		http.Error(w, "Forbidden: Access to this stream is restricted", http.StatusForbidden)
		return
	}

	// Decode JSON payload
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil || len(data) == 0 {
		http.Error(w, "Invalid or missing JSON data", http.StatusBadRequest)
		return
	}

	// Send data asynchronously to Kafka to avoid blocking
	go func() {
		if err := kafka.SendToKafka(streamID, data); err != nil {
			log.Printf("Kafka error for stream %s: %v", streamID, err)
		} else {
			log.Printf("Data sent to Kafka for stream %s", streamID)
		}
	}()

	// Respond immediately
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(SendDataResponse{Status: "data accepted"})
}

// GetResults retrieves and returns results for a stream
func GetResults(w http.ResponseWriter, r *http.Request) {
	if !ValidateAPIKey(r) {
		http.Error(w, "Unauthorized: invalid or missing API key", http.StatusUnauthorized)
		return
	}

	// Extract stream ID from URL
	vars := mux.Vars(r)
	streamID := vars["stream_id"]
	if clientStreamID := r.Header.Get("X-Stream-ID"); clientStreamID != streamID {
		http.Error(w, "Forbidden: Access to this stream is restricted", http.StatusForbidden)
		return
	}

	// Use buffered channel to handle results
	resultChan := make(chan string, 5)
	defer close(resultChan)

	// Process messages asynchronously
	go kafka.ProcessMessages(streamID, resultChan)

	// Set timeout for waiting for results
	select {
	case result := <-resultChan:
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(result)); err != nil {
			log.Printf("Error sending response for stream %s: %v", streamID, err)
		}
	case <-time.After(5 * time.Second):
		http.Error(w, "No data processed within timeout", http.StatusGatewayTimeout)
	}
}

// WebSocket upgrader setup
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// StreamResults streams processed results over WebSocket
func StreamResults(w http.ResponseWriter, r *http.Request) {
	clientAPIKey := r.Header.Get("X-API-Key")
	if clientAPIKey == "" {
		clientAPIKey = r.URL.Query().Get("X-API-Key")
	}

	streamID := r.Header.Get("X-Stream-ID")
	if streamID == "" {
		streamID = r.URL.Query().Get("X-Stream-ID")
	}

	// Validate API key and stream ID
	if clientAPIKey != cachedAPIKey || streamID == "" {
		http.Error(w, "Unauthorized: invalid or missing API key", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP to WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	// Channel for processing results
	resultChan := make(chan string, 5)
	defer close(resultChan)

	go kafka.ProcessMessages(streamID, resultChan)

	// Stream results to WebSocket client
	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, []byte(result)); err != nil {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}
