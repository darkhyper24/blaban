package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration settings
type Config struct {
	MongoURI    string
	RedisAddr   string
	KafkaBroker string
	JWTSecret   string
	Port        string
}

// LoadConfig loads the configuration from environment variables or .env file
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load() // Ignore error if .env doesn't exist

	return &Config{
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017/blaban"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		KafkaBroker: getEnv("KAFKA_BROKER", "localhost:9092"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		Port:        getEnv("PORT", "8080"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
