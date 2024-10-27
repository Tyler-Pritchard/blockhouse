package middleware

import (
	"blockhouse/api/handlers"
	"blockhouse/config"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
)

// Prometheus Metrics
var (
	// Tracks duration of HTTP requests
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	// Counts the total rate limit denials
	rateLimitDenials = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_rate_limit_denials_total",
			Help: "Total number of rate-limited requests denied",
		},
	)
)

func init() {
	prometheus.MustRegister(requestDuration, rateLimitDenials) // Register Prometheus metrics
}

// AuthMiddleware validates API key using handlers' validation function
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !handlers.ValidateAPIKey(r) {
			http.Error(w, "Unauthorized: invalid or missing API key", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs each request with method, path, status, and duration
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		log.Printf("Request received: %s %s", r.Method, r.URL.Path)

		// Wrap ResponseWriter to capture status and response time
		ww := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(ww, r)

		// Calculate and log duration, then record in Prometheus metrics
		duration := time.Since(startTime).Seconds()
		log.Printf("Request completed: %s %s - Status: %d, Duration: %.5fs",
			r.Method, r.URL.Path, ww.statusCode, duration)
		requestDuration.WithLabelValues(r.URL.Path, r.Method).Observe(duration)
	})
}

// statusWriter wraps http.ResponseWriter to capture HTTP status code for logging
type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.statusCode = code
	sw.ResponseWriter.WriteHeader(code)
}

// APIKeyAuthMiddleware verifies the API key from the request header
func APIKeyAuthMiddleware(next http.Handler) http.Handler {
	var (
		once         sync.Once
		cachedAPIKey string
	)

	// Load and cache API key once
	loadAPIKey := func() {
		cachedAPIKey = config.GetAPIKey()
		if cachedAPIKey == "" {
			log.Println("Warning: API_KEY is not set in the environment.")
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(loadAPIKey)
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != cachedAPIKey {
			log.Println("Unauthorized request: invalid or missing API key")
			http.Error(w, `{"error": "Unauthorized: invalid or missing API key"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RateLimiter manages rate limiting for each client
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter initializes a new RateLimiter instance with specified rate and burst limit
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// getLimiter retrieves or creates a rate limiter for a specific client IP
func (rl *RateLimiter) getLimiter(clientIP string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[clientIP]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[clientIP] = limiter
	}
	return limiter
}

// RateLimitMiddleware applies rate limiting based on client IP and increments denial count on limit exceeded
func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.RemoteAddr // Adjust for proxy if needed
			limiter := rl.getLimiter(clientIP)

			if !limiter.Allow() {
				rateLimitDenials.Inc()
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
