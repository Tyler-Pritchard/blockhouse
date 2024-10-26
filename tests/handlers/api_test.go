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

// Load environment variables before each test
func loadEnv(t *testing.T) {
	absPath, err := filepath.Abs("../../.env")
	if err != nil {
		t.Fatalf("Failed to determine absolute path for .env file: %v", err)
	}

	err = godotenv.Load(absPath)
	if err != nil {
		t.Fatalf("Error loading .env file from %s: %v", absPath, err)
	}
	fmt.Println("Loaded environment variables:", os.Getenv("API_KEY"))
}

func TestStartStreamHandler(t *testing.T) {
	loadEnv(t)

	req, err := http.NewRequest("POST", "/stream/start", nil)
	assert.Nil(t, err, "Expected no error creating the request")

	// Add API Key header for authorization
	req.Header.Set("X-API-Key", os.Getenv("API_KEY"))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.StartStream)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "Expected 201 status code")
	assert.Contains(t, rr.Body.String(), "stream_id", "Expected response to contain stream_id")
}

func TestSendDataHandler(t *testing.T) {
	loadEnv(t)

	router := api.SetupRoutes()

	// Prepare JSON payload for the request
	payload := map[string]interface{}{"key": "value"}
	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "/stream/test-stream-id/send", bytes.NewBuffer(payloadBytes))
	assert.Nil(t, err, "Expected no error creating the request")

	// Set headers for API Key and Stream ID
	req.Header.Set("X-API-Key", os.Getenv("API_KEY"))
	req.Header.Set("X-Stream-ID", "test-stream-id")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code, "Expected 202 status code")
	assert.Contains(t, rr.Body.String(), "data accepted", "Expected response to confirm data acceptance")
}
