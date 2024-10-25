package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

// CreateTopic creates a Kafka topic if it doesn't already exist
func CreateTopic(broker string, topic string, numPartitions int) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka broker: %w", err)
	}
	defer conn.Close()

	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: 1,
	})
	if err != nil {
		return fmt.Errorf("failed to create Kafka topic: %w", err)
	}
	return nil
}

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

// SendToKafka sends data to the specified Kafka topic
func SendToKafka(streamID string, data map[string]interface{}) error {
	broker := "localhost:9092" // Kafka broker address
	topic := streamID

	// Ensure the topic exists
	if err := CreateTopic(broker, topic, 1); err != nil {
		return err
	}

	// Initialize Kafka writer with the topic set as streamID
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   streamID,
	})

	// Convert data to JSON
	message, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Write the message to Kafka
	err = writer.WriteMessages(context.Background(), kafka.Message{
		Value: message,
	})
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	// Close the writer
	err = writer.Close()
	if err != nil {
		log.Println("Warning: failed to close Kafka writer:", err)
	}
	return nil
}
