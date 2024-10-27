#!/bin/bash

# ===============================
# Benchmark Script for API Performance
# ===============================

# Configuration Variables
BASE_URL="http://localhost:8080/stream"
CONCURRENT_CONNECTIONS=1000
DURATION=60s
STREAM_ID="your-test-stream-id" # Replace with a valid stream ID for realistic testing

# File paths for WRK scripts
CREATE_STREAM_SCRIPT="benchmark/create_stream.lua"
SEND_DATA_SCRIPT="benchmark/send_data.lua"

# Dependency check for 'wrk'
if ! command -v wrk &> /dev/null; then
  echo "Error: 'wrk' is not installed. Please install it to proceed."
  exit 1
fi

# Benchmarking function for creating a new stream
benchmark_create_stream() {
  echo "Starting benchmark for stream creation..."
  wrk -t12 -c"$CONCURRENT_CONNECTIONS" -d"$DURATION" -s "$CREATE_STREAM_SCRIPT" "$BASE_URL/start" || {
    echo "Error: Benchmark for stream creation failed."
    exit 1
  }
}

# Benchmarking function for sending data to a stream
benchmark_send_data() {
  local send_data_url="$BASE_URL/$STREAM_ID/send"
  echo "Starting benchmark for data sending to stream $STREAM_ID..."
  wrk -t12 -c"$CONCURRENT_CONNECTIONS" -d"$DURATION" -s "$SEND_DATA_SCRIPT" "$send_data_url" || {
    echo "Error: Benchmark for data sending failed."
    exit 1
  }
}

# Main Execution
echo "=============================="
echo "API Benchmark Script Starting"
echo "=============================="
benchmark_create_stream
benchmark_send_data
echo "=============================="
echo "API Benchmarking Complete"
echo "=============================="
