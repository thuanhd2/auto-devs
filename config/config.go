package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host        string
	Port        string
	Username    string
	Password    string
	Name        string
	SSLMode     string
	TablePrefix string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8098"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Host:        getEnv("DB_HOST", "localhost"),
			Port:        getEnv("DB_PORT", "5432"),
			Username:    getEnv("DB_USERNAME", "postgres"),
			Password:    getEnv("DB_PASSWORD", ""),
			Name:        getEnv("DB_NAME", "autodevs"),
			SSLMode:     getEnv("DB_SSLMODE", "disable"),
			TablePrefix: getEnv("DB_TABLE_PREFIX", "dax_"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
