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

// ProcessMessages consumes messages from a Kafka topic, processes them, and sends transformed data through a channel
func ProcessMessages(streamID string, resultChan chan<- string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   streamID,
		GroupID: "consumer-group-" + streamID,
	})
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Error closing Kafka reader for topic %s: %v", streamID, err)
		}
	}()

	messageCounter := 0

	log.Printf("Started consumer for topic %s", streamID)

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message from Kafka for topic %s: %v", streamID, err)
			break
		}

		messageCounter++
		transformedMessage := fmt.Sprintf(
			"Message #%d - Processed at %s: %s",
			messageCounter,
			time.Now().Format(time.RFC3339),
			string(msg.Value),
		)

		log.Printf("Successfully consumed message from topic %s: %s", streamID, transformedMessage)
		resultChan <- transformedMessage
	}
}
