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
	upgrader     = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

// loadAPIKey caches the API key from environment at startup
func loadAPIKey() {
	cachedAPIKey = os.Getenv("API_KEY")
	if cachedAPIKey == "" {
		log.Println("Warning: API_KEY is not set in the environment.")
	}
}

// ValidateAPIKey checks if the request contains a valid API key
func ValidateAPIKey(r *http.Request) bool {
	once.Do(loadAPIKey)
	clientAPIKey := r.Header.Get("X-API-Key")
	return clientAPIKey == cachedAPIKey
}

// StreamResponse represents the response structure when creating a new stream
type StreamResponse struct {
	StreamID string `json:"stream_id"`
}

// StartStream initializes a new data stream and returns a unique stream ID
func StartStream(w http.ResponseWriter, r *http.Request) {
	if !ValidateAPIKey(r) {
		http.Error(w, "Unauthorized: invalid or missing API key", http.StatusUnauthorized)
		return
	}

	streamID := uuid.New().String()
	response := StreamResponse{StreamID: streamID}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("Error encoding StartStream response: %v", err)
	}
}

// SendDataResponse represents the response structure for data sent to a stream
type SendDataResponse struct {
	Status string `json:"status"`
}

// SendData sends a JSON payload to the specified Kafka stream
func SendData(w http.ResponseWriter, r *http.Request) {
	if !ValidateAPIKey(r) {
		http.Error(w, "Unauthorized: invalid or missing API key", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	streamID := vars["stream_id"]
	if clientStreamID := r.Header.Get("X-Stream-ID"); clientStreamID != streamID {
		http.Error(w, "Forbidden: Access to this stream is restricted", http.StatusForbidden)
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil || len(data) == 0 {
		http.Error(w, "Invalid or missing JSON data", http.StatusBadRequest)
		log.Printf("Invalid JSON data for stream %s: %v", streamID, err)
		return
	}

	go func() {
		if err := kafka.SendToKafka(streamID, data); err != nil {
			log.Printf("Failed to send data to Kafka for stream %s: %v", streamID, err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(SendDataResponse{Status: "data accepted"})
}

// GetResults retrieves and sends results for a specified stream
func GetResults(w http.ResponseWriter, r *http.Request) {
	if !ValidateAPIKey(r) {
		http.Error(w, "Unauthorized: invalid or missing API key", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	streamID := vars["stream_id"]
	if clientStreamID := r.Header.Get("X-Stream-ID"); clientStreamID != streamID {
		http.Error(w, "Forbidden: Access to this stream is restricted", http.StatusForbidden)
		return
	}

	resultChan := make(chan string, 5)
	defer close(resultChan)

	go kafka.ProcessMessages(streamID, resultChan)

	select {
	case result := <-resultChan:
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(result)); err != nil {
			log.Printf("Error writing GetResults response for stream %s: %v", streamID, err)
		}
	case <-time.After(5 * time.Second):
		http.Error(w, "No data processed within timeout", http.StatusGatewayTimeout)
		log.Printf("Timeout while retrieving results for stream %s", streamID)
	}
}

// StreamResults establishes a WebSocket connection for streaming Kafka results
func StreamResults(w http.ResponseWriter, r *http.Request) {
	clientAPIKey := r.Header.Get("X-API-Key")
	if clientAPIKey == "" {
		clientAPIKey = r.URL.Query().Get("X-API-Key")
	}

	streamID := r.Header.Get("X-Stream-ID")
	if streamID == "" {
		streamID = r.URL.Query().Get("X-Stream-ID")
	}

	if clientAPIKey != cachedAPIKey || streamID == "" {
		http.Error(w, "Unauthorized: invalid or missing API key", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed for stream %s: %v", streamID, err)
		return
	}
	defer conn.Close()

	resultChan := make(chan string, 5)
	defer close(resultChan)

	go kafka.ProcessMessages(streamID, resultChan)

	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, []byte(result)); err != nil {
				log.Printf("WebSocket send error for stream %s: %v", streamID, err)
				return
			}
		case <-r.Context().Done():
			log.Printf("Client disconnected from WebSocket for stream %s", streamID)
			return
		}
	}
}
