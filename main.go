package main

import (
	"blockhouse/api"
	"blockhouse/api/middleware"
	"blockhouse/config"
	"blockhouse/kafka"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of requests to the API",
		},
		[]string{"endpoint", "method"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_request_duration_seconds",
			Help:    "Duration of API requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)
	kafkaMessageCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_messages_total",
			Help: "Total number of messages processed by Kafka",
		},
	)
)

func init() {
	prometheus.MustRegister(requestCount, requestDuration, kafkaMessageCount)
}

func main() {
	// Load environment variables
	config.LoadEnv()

	// Set up routes
	router := api.SetupRoutes()

	// Add Prometheus metrics handler
	router.Handle("/metrics", promhttp.Handler())

	// Initialize rate limiter: 10 requests per second with a burst capacity of 20
	rateLimiter := middleware.NewRateLimiter(10, 20)

	// Apply middleware in order
	router.Use(middleware.LoggingMiddleware)                // Logs all requests
	router.Use(middleware.AuthMiddleware)                   // Validates the API key
	router.Use(middleware.RateLimitMiddleware(rateLimiter)) // Throttles excessive requests

	// Start the server
	port := config.GetEnv("WEBSOCKET_PORT")
	log.Println("Server is starting on port", port)
	log.Fatal(http.ListenAndServe(":8080", router))

	// Initialize Kafka topic
	topic := "stream_topic"

	// Start the consumer in a goroutine to continuously consume messages
	go kafka.ConsumeMessages(topic)

	// Sleep for a few seconds to allow the consumer to start
	time.Sleep(2 * time.Second)

	// Produce a message
	kafka.ProduceMessage(topic, "Hello Redpanda!")
}
