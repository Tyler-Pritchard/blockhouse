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

var (
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	rateLimitDenials = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "rate_limit_denials_total",
			Help: "Total number of rate-limited requests denied",
		},
	)
)

func init() {
	// Register Prometheus metrics
	prometheus.MustRegister(requestDuration, rateLimitDenials)
}

// AuthMiddleware uses the handlers' API key validation
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Call ValidateAPIKey from handlers
		if !handlers.ValidateAPIKey(r) {
			http.Error(w, "Unauthorized: invalid or missing API key", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs each incoming request with method, path, and response time.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Log the incoming request with method and URL path
		log.Printf("Request received: %s %s", r.Method, r.URL.Path)

		// Wrap ResponseWriter to capture status and response time
		ww := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(ww, r)

		// Log the completion of the request with method, path, and duration
		duration := time.Since(startTime).Seconds()
		log.Printf("Request completed: %s %s - Status: %d, Duration: %v",
			r.Method, r.URL.Path, ww.statusCode, duration)

		// Record the duration to Prometheus metrics
		requestDuration.WithLabelValues(r.URL.Path, r.Method).Observe(duration)
	})
}

// statusWriter wraps http.ResponseWriter to capture the HTTP status code
type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.statusCode = code
	sw.ResponseWriter.WriteHeader(code)
}

// Cached API key for efficiency
var cachedAPIKey string
var once sync.Once

// loadAPIKey caches the API key from config at startup
func loadAPIKey() {
	cachedAPIKey = config.GetAPIKey()
	if cachedAPIKey == "" {
		log.Println("Warning: API_KEY is not set in the environment.")
	}
}

// APIKeyAuthMiddleware validates the API key in the request header
func APIKeyAuthMiddleware(next http.Handler) http.Handler {
	once.Do(loadAPIKey) // Ensures the API key is loaded only once

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != cachedAPIKey {
			log.Println("Unauthorized request: invalid or missing API key")
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error": "Unauthorized: invalid or missing API key"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RateLimiter holds a rate limiter for each client
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter initializes a new RateLimiter
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// getLimiter retrieves or creates a rate limiter for a specific client
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

// RateLimitMiddleware applies rate limiting to each client based on their IP address
func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.RemoteAddr // Extract IP from headers if behind a proxy
			limiter := rl.getLimiter(clientIP)

			// Deny request if rate limit exceeded
			if !limiter.Allow() {
				rateLimitDenials.Inc() // Increment the rate limit denial counter
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
