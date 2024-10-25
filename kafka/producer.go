package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

func ProduceMessage(topic string, message string) {
	writer := kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"), // Default Redpanda port
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte("key"),
			Value: []byte(message),
		},
	)

	if err != nil {
		log.Fatal("Failed to write message:", err)
	}

	log.Println("Message sent:", message)
	writer.Close()
}
