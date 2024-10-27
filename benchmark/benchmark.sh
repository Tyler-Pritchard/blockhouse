#!/bin/bash

# Variables for configuration
URL="http://localhost:8080/stream/start"  # Adjust to point to your API endpoint
CONCURRENT_CONNECTIONS=1000
DURATION=60s

# Function to benchmark stream creation
echo "Benchmarking stream creation..."
wrk -t12 -c$CONCURRENT_CONNECTIONS -d$DURATION -s benchmark/create_stream.lua $URL

# Function to benchmark data sending to a stream
STREAM_ID="your-test-stream-id" # Replace with a valid stream ID for testing data load
SEND_DATA_URL="http://localhost:8080/stream/$STREAM_ID/send"
echo "Benchmarking data sending..."
wrk -t12 -c$CONCURRENT_CONNECTIONS -d$DURATION -s benchmark/send_data.lua $SEND_DATA_URL
