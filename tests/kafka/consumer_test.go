package kafka_test

import (
	"blockhouse/kafka"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestProcessMessages verifies that messages from Kafka are processed as expected
func TestProcessMessages(t *testing.T) {
	streamID := "test-stream-id"
	resultChan := make(chan string, 1) // Buffered channel for message processing

	// Start Kafka message processing in a goroutine
	go kafka.ProcessMessages(streamID, resultChan)

	// Wait for a message in resultChan or time out after 3 seconds
	select {
	case result := <-resultChan:
		assert.Contains(t, result, "Processed", "Expected message to be in processed format")
	case <-time.After(3 * time.Second):
		t.Fatal("Test timed out: message processing took too long")
	}
}
