package kafka_test

import (
	"context"
	"encoding/json"
	"testing"

	kafkalib "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

// TestSendToKafka verifies that a message is successfully sent to Kafka
func TestSendToKafka(t *testing.T) {
	streamID := "test-stream-id"
	data := map[string]interface{}{"key": "value"}

	// Setup Kafka writer with mock configuration for testing
	writer := kafkalib.NewWriter(kafkalib.WriterConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    streamID,
		Balancer: &kafkalib.LeastBytes{},
	})
	defer func() {
		if err := writer.Close(); err != nil {
			t.Logf("Warning: Failed to close Kafka writer: %v", err)
		}
	}()

	// Convert data to JSON format for Kafka message
	message, err := json.Marshal(data)
	assert.Nil(t, err, "Expected no error when marshalling data to JSON")

	// Send message to Kafka and check for errors
	err = writer.WriteMessages(context.Background(), kafkalib.Message{
		Key:   []byte(streamID),
		Value: message,
	})
	assert.Nil(t, err, "Expected no error when sending message to Kafka")
}
