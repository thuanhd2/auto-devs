package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Worktree   WorktreeConfig
	Redis      RedisConfig
	Centrifuge CentrifugeConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Name     string
	SSLMode  string
}

type WorktreeConfig struct {
	BaseDirectory   string
	MaxPathLength   int
	MinDiskSpace    int64
	CleanupInterval string
	EnableLogging   bool
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type CentrifugeConfig struct {
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	Engine        string // "redis" or "memory"
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
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Username: getEnv("DB_USERNAME", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "autodevs"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Worktree: WorktreeConfig{
			BaseDirectory:   getEnv("WORKTREE_BASE_DIR", "/worktrees"),
			MaxPathLength:   getEnvAsInt("WORKTREE_MAX_PATH_LENGTH", 4096),
			MinDiskSpace:    getEnvAsInt64("WORKTREE_MIN_DISK_SPACE", 100*1024*1024), // 100MB
			CleanupInterval: getEnv("WORKTREE_CLEANUP_INTERVAL", "24h"),
			EnableLogging:   getEnvAsBool("WORKTREE_ENABLE_LOGGING", true),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Centrifuge: CentrifugeConfig{
			RedisHost:     getEnv("CENTRIFUGE_REDIS_HOST", "localhost"),
			RedisPort:     getEnv("CENTRIFUGE_REDIS_PORT", "6379"),
			RedisPassword: getEnv("CENTRIFUGE_REDIS_PASSWORD", ""),
			RedisDB:       getEnvAsInt("CENTRIFUGE_REDIS_DB", 1), // Use different DB than main Redis
			Engine:        getEnv("CENTRIFUGE_ENGINE", "redis"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
