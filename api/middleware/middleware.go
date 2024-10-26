package middleware

import (
	"blockhouse/config"
	"log"
	"net/http"
)

// AuthMiddleware is a placeholder middleware for API authentication
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Placeholder for authentication logic
		// For now, it just calls the next handler
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware is a placeholder middleware for logging requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request received:", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// APIKeyAuthMiddleware checks for a valid API key in the request header
func APIKeyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" || apiKey != config.GetAPIKey() {
			http.Error(w, "Unauthorized: invalid or missing API key", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
