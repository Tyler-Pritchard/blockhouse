package kafka_test

import (
	"blockhouse/kafka"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestProcessMessages verifies message processing from Kafka
func TestProcessMessages(t *testing.T) {
	streamID := "test-stream-id"
	resultChan := make(chan string, 1)

	// Simulate Kafka message
	go func() {
		kafka.ProcessMessages(streamID, resultChan)
	}()

	select {
	case result := <-resultChan:
		assert.Contains(t, result, "Processed", "Expected processed message format")
	case <-time.After(3 * time.Second):
		t.Error("Timeout: message processing took too long")
	}
}
