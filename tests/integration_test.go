package main

import (
	"blockhouse/api"
	"blockhouse/api/handlers"
	"blockhouse/kafka"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// Load environment variables for integration tests
func loadEnvForIntegrationTests(t *testing.T) {
	err := godotenv.Load("../.env") // Adjust path if necessary
	if err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}
}

// Helper function to create a new stream and return its ID
func createStream(t *testing.T) string {
	loadEnvForIntegrationTests(t)

	// Initialize the router
	router := api.SetupRoutes()

	// Start a new stream
	req, err := http.NewRequest("POST", "/stream/start", nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("X-API-Key", os.Getenv("API_KEY"))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Validate stream creation
	if rr.Code != http.StatusCreated {
		t.Fatalf("Failed to start stream, got status code: %v", rr.Code)
	}
	var response map[string]string
	json.NewDecoder(rr.Body).Decode(&response)
	return response["stream_id"]
}

func TestIntegrationStreamingData(t *testing.T) {
	// Use createStream to set up a valid stream ID
	streamID := createStream(t)

	// Initialize the router
	router := api.SetupRoutes()

	// Step 2: Send data to the stream
	payload := map[string]interface{}{"key": "integration-test-value"}
	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "/stream/"+streamID+"/send", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("Error creating send data request: %v", err)
	}
	req.Header.Set("X-API-Key", os.Getenv("API_KEY"))
	req.Header.Set("X-Stream-ID", streamID)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check that the message was accepted
	if rr.Code != http.StatusAccepted {
		t.Fatalf("Failed to send data, got status code: %v", rr.Code)
	}

	// Step 3: Validate that Kafka received and processed the data
	// Set up a Kafka consumer to check for the message
}

func TestWebSocketStreaming(t *testing.T) {
	loadEnvForIntegrationTests(t)

	// Use createStream to get a valid stream ID
	streamID := createStream(t)

	// Set up WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Call handler with proper headers
		handlers.StreamResults(w, r)
	}))
	defer server.Close()

	// Construct WebSocket URL with API Key and Stream ID as query parameters
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/" + streamID + "?X-API-Key=" + os.Getenv("API_KEY") + "&X-Stream-ID=" + streamID

	// Establish WebSocket connection using URL with query params
	wsConn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to establish WebSocket connection: %v, HTTP status: %d", err, resp.StatusCode)
	}
	defer wsConn.Close()

	// Simulate Kafka producer sending a message
	go func() {
		kafka.SendToKafka(streamID, map[string]interface{}{"key": "integration-test-value"})
	}()

	// Wait for WebSocket message
	var msg []byte
	wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err = wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("Error reading WebSocket message: %v", err)
	}

	// Validate received message
	expectedMessage := `{"key":"integration-test-value"}`
	if !bytes.Contains(msg, []byte(expectedMessage)) {
		t.Errorf("Expected WebSocket message to contain: %s, but got: %s", expectedMessage, msg)
	}
}
