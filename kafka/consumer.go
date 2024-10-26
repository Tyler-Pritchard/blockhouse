package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func ConsumeMessages(topic string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   topic,
		GroupID: "my-group",
	})

	for {
		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatal("Failed to read message:", err)
		}

		log.Printf("Message received: %s", string(message.Value))
	}
}

// ProcessMessages consumes messages from a Kafka topic and sends transformed data through a channel
func ProcessMessages(streamID string, resultChan chan<- string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   streamID,
		GroupID: "consumer-group-" + streamID, // Ensures each stream has its own consumer group
	})

	defer reader.Close()

	// Listen for messages and process them
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message from Kafka: %v\n", err)
			break
		}

		// Perform a simple transformation (e.g., adding a timestamp)
		transformedMessage := fmt.Sprintf("Processed at %s: %s", time.Now().Format(time.RFC3339), string(msg.Value))
		log.Printf("Consumed message from stream %s: %s", streamID, transformedMessage)

		// Send the processed message to the result channel
		resultChan <- transformedMessage
	}
}
