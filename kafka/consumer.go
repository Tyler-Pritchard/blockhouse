package kafka

import (
	"context"
	"log"

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
