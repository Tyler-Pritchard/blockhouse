package handlers_test

import (
	"blockhouse/api"
	"blockhouse/api/handlers"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// loadEnv loads environment variables from .env file for testing
func loadEnv(t *testing.T) {
	t.Helper() // Marks this function as a helper to improve test output readability

	absPath, err := filepath.Abs("../../.env")
	if err != nil {
		t.Fatalf("Failed to determine absolute path for .env file: %v", err)
	}

	if err = godotenv.Load(absPath); err != nil {
		t.Fatalf("Error loading .env file from %s: %v", absPath, err)
	}
	fmt.Printf("Loaded environment variable API_KEY: %s\n", os.Getenv("API_KEY"))
}

// TestStartStreamHandler validates the creation of a new stream
func TestStartStreamHandler(t *testing.T) {
	loadEnv(t)

	req, err := http.NewRequest(http.MethodPost, "/stream/start", nil)
	assert.NoError(t, err, "Failed to create POST request for /stream/start")

	req.Header.Set("X-API-Key", os.Getenv("API_KEY")) // Set API Key for authorization

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.StartStream)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "Expected 201 status code for stream creation")
	assert.Contains(t, rr.Body.String(), "stream_id", "Response should contain stream_id field")
}

// TestSendDataHandler validates sending data to an existing stream
func TestSendDataHandler(t *testing.T) {
	loadEnv(t)
	router := api.SetupRoutes()

	// Prepare JSON payload
	payload := map[string]interface{}{"key": "value"}
	payloadBytes, err := json.Marshal(payload)
	assert.NoError(t, err, "Failed to marshal payload for /stream/{stream_id}/send")

	req, err := http.NewRequest(http.MethodPost, "/stream/test-stream-id/send", bytes.NewBuffer(payloadBytes))
	assert.NoError(t, err, "Failed to create POST request for /stream/{stream_id}/send")

	// Set necessary headers
	req.Header.Set("X-API-Key", os.Getenv("API_KEY"))
	req.Header.Set("X-Stream-ID", "test-stream-id")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code, "Expected 202 status code for data sending")
	assert.Contains(t, rr.Body.String(), "data accepted", "Response should confirm data acceptance")
}
