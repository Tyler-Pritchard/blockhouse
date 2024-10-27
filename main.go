package main

import (
	"blockhouse/api"
	"blockhouse/api/middleware"
	"blockhouse/config"
	"blockhouse/kafka"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Define Prometheus metrics
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
	// Register metrics for monitoring
	prometheus.MustRegister(requestCount, requestDuration, kafkaMessageCount)
}

func main() {
	// Load environment configuration
	config.LoadEnv()

	// Set up router with middleware and routes
	router := setupRouter()

	// Start Kafka consumer in a separate goroutine
	initializeKafkaConsumer("stream_topic")

	// Start the HTTP server
	port := config.GetEnv("WEBSOCKET_PORT")
	log.Printf("Server is starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// setupRouter configures API routes, applies middleware, and sets up the Prometheus endpoint
func setupRouter() *mux.Router {
	router := api.SetupRoutes()

	// Add Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler())

	// Initialize rate limiter with a limit of 10 requests per second and burst capacity of 20
	rateLimiter := middleware.NewRateLimiter(10, 20)

	// Apply middlewares in the preferred order
	router.Use(
		middleware.LoggingMiddleware,                // Logs incoming requests
		middleware.AuthMiddleware,                   // Validates the API key
		middleware.RateLimitMiddleware(rateLimiter), // Throttles excessive requests
	)

	return router
}

// initializeKafkaConsumer starts a Kafka consumer to process messages in a separate goroutine
func initializeKafkaConsumer(topic string) {
	go kafka.ConsumeMessages(topic) // Start Kafka consumer

	// Allow the consumer to initialize before producing test messages
	time.Sleep(2 * time.Second)

	// Produce a sample message for testing
	kafka.ProduceMessage(topic, "Hello Redpanda!")
}
