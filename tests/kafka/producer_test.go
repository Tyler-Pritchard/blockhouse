package kafka_test

import (
	"context"
	"encoding/json"
	"testing"

	kafkalib "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

// TestSendToKafka verifies message sending to Kafka
func TestSendToKafka(t *testing.T) {
	streamID := "test-stream-id"
	data := map[string]interface{}{"key": "value"}

	// Mock Kafka writer
	writer := kafkalib.NewWriter(kafkalib.WriterConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    streamID,
		Balancer: &kafkalib.LeastBytes{},
	})

	// Convert data to JSON format
	message, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}

	// Write message to Kafka (mocked)
	err = writer.WriteMessages(context.Background(), kafkalib.Message{
		Key:   []byte(streamID),
		Value: message,
	})
	assert.Nil(t, err, "Expected no error when sending message to Kafka")
	writer.Close()
}
