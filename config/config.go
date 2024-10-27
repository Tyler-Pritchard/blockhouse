package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file, logging an error if unsuccessful.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

// GetEnv retrieves an environment variable by key.
// Returns an empty string if the key does not exist.
func GetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("Warning: Environment variable %s is not set or empty", key)
	}
	return value
}

// GetAPIKey is a convenience function to retrieve the API key specifically.
func GetAPIKey() string {
	return GetEnv("API_KEY")
}
