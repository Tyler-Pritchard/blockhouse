package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// ConsumeMessages reads messages continuously from the specified Kafka topic
// and logs each message received.
func ConsumeMessages(topic string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   topic,
		GroupID: "my-group",
	})
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Error closing Kafka reader for topic %s: %v", topic, err)
		}
	}()

	for {
		// Read a message from the Kafka topic
		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatalf("Failed to read message from Kafka topic %s: %v", topic, err)
		}

		log.Printf("Message received from topic %s: %s", topic, string(message.Value))
	}
}

// ProcessMessages consumes messages from a Kafka topic, processes each message,
// and sends the transformed data through a result channel.
func ProcessMessages(streamID string, resultChan chan<- string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   streamID,
		GroupID: "consumer-group-" + streamID,
	})
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Error closing Kafka reader for stream %s: %v", streamID, err)
		}
	}()

	messageCounter := 0
	log.Printf("Started message processing for stream %s", streamID)

	for {
		// Read a message from the Kafka topic for the given stream
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message from Kafka for stream %s: %v", streamID, err)
			break
		}

		messageCounter++
		transformedMessage := formatMessage(messageCounter, msg.Value)

		log.Printf("Processed message from stream %s: %s", streamID, transformedMessage)
		resultChan <- transformedMessage
	}
}

// formatMessage applies consistent formatting to a Kafka message for logging and channel transmission.
func formatMessage(counter int, value []byte) string {
	return fmt.Sprintf(
		"Message #%d - Processed at %s: %s",
		counter,
		time.Now().Format(time.RFC3339),
		string(value),
	)
}
