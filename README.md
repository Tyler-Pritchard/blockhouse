# Real-Time Data Streaming API

A high-performance real-time data streaming API built in Go, leveraging Kafka for message brokering, Prometheus for metrics, and comprehensive middleware for authentication, rate limiting, and logging. This project includes industry-standard benchmarks and metrics to support production-grade performance and monitoring.

## Table of Contents

- [Real-Time Data Streaming API](#real-time-data-streaming-api)
  - [Table of Contents](#table-of-contents)
    - [Project Overview](#project-overview)
    - [Architecture](#architecture)
    - [Features](#features)
    - [Setup \& Installation](#setup--installation)
    - [Environment Variables](#environment-variables)
    - [Benchmarking \& Performance](#benchmarking--performance)
    - [Metrics \& Monitoring](#metrics--monitoring)
    - [Testing](#testing)
    - [Directory Structure](#directory-structure)
    - [Future Enhancements](#future-enhancements)

### Project Overview

This project provides a scalable API for real-time data streaming, suitable for high-throughput applications. It supports Kafka-based messaging, allowing clients to initiate, send data to, and retrieve results from dynamically generated streams. The API includes:

- **Kafka** integration for distributed message handling.
- **Prometheus** for real-time metrics.
- **Middleware** for rate-limiting, logging, and authentication.
- **Benchmarking** to ensure reliability under heavy load.

### Architecture

- **API Layer**: Exposes endpoints for managing streams and sending/receiving data.
- **Middleware**: Implements logging, API key validation, and rate limiting.
- **Kafka Integration**: Manages topic creation, message production, and consumption for scalable data handling.
- **Metrics**: Provides insights into API request times, Kafka message counts, and request limits using Prometheus.

### Features

- **Stream Management**: Create, send data to, and retrieve results from unique Kafka streams.
- **Middleware**:
  - **AuthMiddleware**: Validates requests using API keys.
  - **RateLimitMiddleware**: Controls request rate per client IP.
  - **LoggingMiddleware**: Logs detailed request and response times.
- **Benchmarking**: Scripts for performance testing using WRK, with customizable concurrent connections.
- **Metrics**: Exposes Prometheus-compatible metrics to monitor API and Kafka performance.

### Setup & Installation

**1. Clone the Repository**:
```
git clone https://github.com/your-username/real-time-data-streaming-api.git
cd real-time-data-streaming-api
```

**2. Install Dependencies**: Ensure [Go](https://golang.org/) and [WRK](https://github.com/wg/wrk) (for benchmarking) are installed.

**3. Set up Kafka**:

Install and configure Kafka. Update the `KAFKA_BROKER` in `.env` as needed.

**4. Configure Environment**: Copy `.env.example` to `.env` and adjust settings as necessary.

**5. Start the Application**:
```
go run main.go
```

### Environment Variables

Set environment variables in `.env`:
```
KAFKA_BROKER=localhost:9092           # Kafka broker address
WEBSOCKET_PORT=8080                   # API and WebSocket server port
API_KEY=your_secret_api_key_here      # API Key for authentication
```

### Benchmarking & Performance
The `benchmark/benchmark.sh` script provides automated benchmarking for API performance under load.

**Usage**:
```
chmod +x benchmark/benchmark.sh
./benchmark/benchmark.sh
```

- **Stream Creation**: Benchmarks the /stream/start endpoint.
- **Data Sending**: Benchmarks the /stream/{stream_id}/send endpoint.

Benchmark results are logged, offering insights into latency and throughput under a configurable number of concurrent connections.

**Example Output**:

Refer to example benchmark results in the logs for average latency, request per second, and error rates.

### Metrics & Monitoring

Prometheus metrics are exposed at `/metrics` endpoint:

- **API Request Counts**: Total count of requests per endpoint.
- **Request Duration**: Histograms of request times.
- **Rate Limit Denials**: Counts of requests denied due to rate limits.
- **Kafka Message Metrics**: Kafka-specific metrics like message count and message duration.
  
Integrate with [Grafana](https://grafana.com/) for visualizing metrics.

### Testing

The `tests/` directory includes unit and integration tests.

**Running Tests**:
```
go test ./...
```

- **Unit Tests**: Validate core functionality and individual methods.
- **Integration Tests**: Confirm Kafka and API interactions with a live or test Kafka instance.
  
### Directory Structure
```
/api                    # API route handlers and middleware
/benchmark              # WRK benchmarking scripts
/config                 # Environment and configuration management
/kafka                  # Kafka producer/consumer implementations
/models                 # Data models
/tests                  # Unit and integration tests
.env                    # Environment variable definitions
README.md               # Project documentation
```

### Future Enhancements
- **Advanced Authentication**: Implement OAuth2 for secure access.
- **Enhanced Rate Limiting**: Add global rate limits with Redis.
- **Improved Kafka Error Handling**: Graceful handling and retry logic for Kafka outages.