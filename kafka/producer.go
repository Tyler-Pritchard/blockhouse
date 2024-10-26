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

// ProduceMessage sends a simple message to a specific topic
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
	broker := "localhost:9092"
	topic := streamID

	if err := CreateTopic(broker, topic, 1); err != nil {
		log.Printf("Error creating topic %s: %v", topic, err)
		return fmt.Errorf("failed to create topic: %w", err)
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{broker},
		Topic:    streamID,
		Balancer: &kafka.CRC32Balancer{},
	})

	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshalling data for topic %s: %v", topic, err)
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	log.Printf("Producing message to topic %s: %s", topic, message)

	err = writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(streamID),
		Value: message,
	})
	if err != nil {
		log.Printf("Error writing message to Kafka for topic %s: %v", topic, err)
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	log.Printf("Successfully produced message to topic %s", topic)

	if err = writer.Close(); err != nil {
		log.Printf("Error closing Kafka writer for topic %s: %v", topic, err)
	}

	return nil
}
