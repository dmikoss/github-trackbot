package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

type Config struct {
	TelegramBotToken    string
	TelegramApiHost     string
	GithubFetchRate     float64 // github fetch rate limiting (in RPS)
	GithubFetchTimeout  float64 // github fetch http client timeout
	TelegramHttpTimeout float64 // telegram http client timeout
}

// New returns a new Config struct
func New() *Config {
	return &Config{
		TelegramBotToken:    getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramApiHost:     getEnv("TELEGRAM_API_HOST", "api.telegram.org"),
		GithubFetchRate:     getEnvFloat("GITHUB_FETCH_RATE", 0.5),
		GithubFetchTimeout:  getEnvFloat("GITHUB_FETCH_TIMEOUT", 10),
		TelegramHttpTimeout: getEnvFloat("TELEGRAM_HTTP_TIMEOUT", 10),
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return defaultVal
		}
		return v
	}
	return defaultVal
}
