package config

import (
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	LogLevel       string
	Port           string
	HttpTimeout    string
	AgifyURL       string
	GenderizeURL   string
	NationalizeURL string
}

func LoadConfig(logger *zap.Logger) Config {
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../../.env"); err != nil {
			logger.Fatal("Failed to load .env file", zap.Error(err))
		}
	}

	return Config{
		DBHost:         getOrDefault("DB_HOST", "localhost"),
		DBPort:         getOrDefault("DB_PORT", "5432"),
		DBUser:         getOrDefault("DB_USER", "postgres"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBName:         os.Getenv("DB_NAME"),
		DBSSLMode:      getOrDefault("DB_SSL_MODE", "disable"),
		LogLevel:       getOrDefault("LOG_LEVEL", "debug"),
		Port:           getOrDefault("PORT", "8080"),
		HttpTimeout:    getOrDefault("HTTP_TIMEOUT", "5"),
		AgifyURL:       os.Getenv("API_AGIFY"),
		GenderizeURL:   os.Getenv("API_GENDERIZE"),
		NationalizeURL: os.Getenv("API_NATIONALIZE"),
	}
}

func getOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
