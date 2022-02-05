package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

type Config struct {
	TelegramToken string
	TelegramHost  string
}

// New returns a new Config struct
func New() *Config {
	return &Config{
		TelegramToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramHost:  getEnv("TELEGRAM_API_HOST", "api.telegram.org"),
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
