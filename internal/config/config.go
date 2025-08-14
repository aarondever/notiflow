package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

type Config struct {
	AppEnv      string // Environment type (development, production)
	Host        string
	Port        int
	DatabaseURL string
	DBName      string
	Timezone    *time.Location
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	// Get timezone from environment or use UTC as default
	tzName := getStringEnv("TZ", "UTC")
	timezone, err := time.LoadLocation(tzName)
	if err != nil {
		log.Printf("Invalid timezone: %s, using UTC instead", tzName)
		timezone = time.UTC
	}

	cfg := &Config{
		AppEnv:   getStringEnv("APP_ENV", "development"),
		Host:     getStringEnv("HOST", "0.0.0.0"),
		Port:     getIntEnv("PORT", 8080),
		DBName:   getStringEnv("DB_NAME", "go_notifier"),
		Timezone: timezone,
	}

	// Build database URL
	dbHost := getStringEnv("DB_HOST", "localhost")
	dbPort := getIntEnv("DB_PORT", 27017)
	dbUser := getStringEnv("DB_USER", "mongo")
	dbPassword := getStringEnv("DB_PASSWORD", "mongo")

	cfg.DatabaseURL = fmt.Sprintf("mongodb://%s:%s@%s:%d",
		dbUser, dbPassword, dbHost, dbPort)

	return cfg, nil
}

// getStringEnv retrieves a string environment variable with a default value
func getStringEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv retrieves an integer environment variable with a default value
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		for _, char := range value {
			if char >= '0' && char <= '9' {
				result = result*10 + int(char-'0')
			} else {
				log.Printf("Invalid integer value for %s: %s, using default: %d", key, value, defaultValue)
				return defaultValue
			}
		}
		return result
	}
	return defaultValue
}
