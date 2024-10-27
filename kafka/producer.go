package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	topicConnOnce sync.Once
	topicConn     *kafka.Conn
)

// getKafkaConnection creates or reuses a Kafka connection for topic operations
func getKafkaConnection(broker string) (*kafka.Conn, error) {
	var err error
	topicConnOnce.Do(func() {
		topicConn, err = kafka.Dial("tcp", broker)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Kafka broker: %w", err)
	}
	return topicConn, nil
}

// CreateTopic creates a Kafka topic if it doesn't already exist
func CreateTopic(broker, topic string, numPartitions int) error {
	conn, err := getKafkaConnection(broker)
	if err != nil {
		return fmt.Errorf("failed to obtain Kafka connection: %w", err)
	}

	// Check if the topic already exists to avoid redundant creation attempts
	partitions, err := conn.ReadPartitions(topic)
	if err == nil && len(partitions) > 0 {
		return nil // Topic exists; no need to create
	}

	// Attempt to create the topic if it doesn't exist
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

var (
	writerOnce    sync.Once
	kafkaWriter   *kafka.Writer
	brokerAddress = "localhost:9092" // set broker once
)

// getKafkaWriter initializes or returns a persistent Kafka writer
func getKafkaWriter(topic string) *kafka.Writer {
	writerOnce.Do(func() {
		kafkaWriter = kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{brokerAddress},
			Topic:    topic,
			Balancer: &kafka.CRC32Balancer{},
		})
	})
	return kafkaWriter
}

// ProduceMessage sends a simple message to a specific topic
func ProduceMessage(topic, message string) {
	writer := getKafkaWriter(topic)

	// Use a timeout context to avoid indefinite blocking
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("key"),
		Value: []byte(message),
	})
	if err != nil {
		log.Printf("Error writing message to topic %s: %v", topic, err)
		return
	}

	log.Println("Message sent:", message)
}

// SendToKafka sends data to the specified Kafka topic
func SendToKafka(streamID string, data map[string]interface{}) error {
	topic := streamID
	writer := getKafkaWriter(topic)

	// Marshal data to JSON
	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshalling data for topic %s: %v", topic, err)
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	log.Printf("Producing message to topic %s: %s", topic, message)

	// Use context with timeout to avoid indefinite blocking
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(streamID),
		Value: message,
	})
	if err != nil {
		log.Printf("Error writing message to Kafka for topic %s: %v", topic, err)
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	log.Printf("Successfully produced message to topic %s", topic)
	return nil
}
