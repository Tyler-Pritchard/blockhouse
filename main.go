package main

import (
	"blockhouse/api"
	"blockhouse/config"
	"blockhouse/kafka"
	"log"
	"net/http"
	"time"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Set up routes
	router := api.SetupRoutes()

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
