package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"
)

var (
	// Prometheus metrics for Kafka message tracking
	kafkaMessageCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_message_count",
			Help: "Total number of messages sent to Kafka",
		},
		[]string{"topic"},
	)
	kafkaMessageDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_message_duration_seconds",
			Help:    "Duration of Kafka message production in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topic"},
	)
)

func init() {
	prometheus.MustRegister(kafkaMessageCount, kafkaMessageDuration)
}

var (
	topicConnOnce sync.Once
	topicConn     *kafka.Conn
)

// getKafkaConnection creates or retrieves a Kafka connection for topic operations.
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

// CreateTopic ensures a Kafka topic exists or creates it if missing.
func CreateTopic(broker, topic string, numPartitions int) error {
	conn, err := getKafkaConnection(broker)
	if err != nil {
		return fmt.Errorf("failed to obtain Kafka connection: %w", err)
	}

	// Check for topic existence to avoid redundant creation
	partitions, err := conn.ReadPartitions(topic)
	if err == nil && len(partitions) > 0 {
		return nil
	}

	// Create the topic if it doesn't already exist
	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: 1,
	})
	if err != nil {
		return fmt.Errorf("failed to create Kafka topic: %w", err)
	}
	log.Printf("Topic %s created with %d partitions", topic, numPartitions)
	return nil
}

var (
	writerOnce    sync.Once
	kafkaWriter   *kafka.Writer
	brokerAddress = "localhost:9092" // Default Kafka broker address
)

// getKafkaWriter initializes or reuses a Kafka writer for message production.
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

// ProduceMessage sends a text message to a specific Kafka topic.
func ProduceMessage(topic, message string) {
	writer := getKafkaWriter(topic)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	startTime := time.Now()

	if err := writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("key"),
		Value: []byte(message),
	}); err != nil {
		log.Printf("Failed to write message to topic %s: %v", topic, err)
		return
	}

	// Log and record metrics
	logMessageMetrics(topic, time.Since(startTime).Seconds())
	log.Printf("Message sent to topic %s: %s", topic, message)
}

// SendToKafka marshals and sends structured data to the specified Kafka topic.
func SendToKafka(streamID string, data map[string]interface{}) error {
	topic := streamID
	writer := getKafkaWriter(topic)

	// Marshal data to JSON
	message, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data for topic %s: %w", topic, err)
	}
	log.Printf("Preparing message for topic %s: %s", topic, message)

	// Set timeout and start timer for metrics
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	startTime := time.Now()

	// Send message to Kafka
	if err := writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(streamID),
		Value: message,
	}); err != nil {
		return fmt.Errorf("failed to write message to Kafka for topic %s: %w", topic, err)
	}

	// Log and record metrics
	logMessageMetrics(topic, time.Since(startTime).Seconds())
	log.Printf("Message successfully sent to topic %s", topic)
	return nil
}

// logMessageMetrics records message metrics to Prometheus.
func logMessageMetrics(topic string, duration float64) {
	kafkaMessageCount.WithLabelValues(topic).Inc()
	kafkaMessageDuration.WithLabelValues(topic).Observe(duration)
}
