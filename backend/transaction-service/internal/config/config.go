package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config содержит все настройки сервиса
type Config struct {
	// Server
	GRPCPort string

	// PostgreSQL
	PostgresDSN string

	// Redis
	RedisAddr     string
	RedisPassword string

	// Kafka
	KafkaBrokers []string
	KafkaTopic   string

	// ClickHouse (будет позже)
	ClickHouseDSN string

	// JWT (если понадобится)
	JWTSecret         string
	JWTExpiration     time.Duration
	RefreshExpiration time.Duration
}

// Load загружает конфигурацию из переменных окружения или использует значения по умолчанию
func Load() *Config {
	return &Config{
		GRPCPort:          getEnv("GRPC_PORT", "50052"),
		PostgresDSN:       getEnv("POSTGRES_DSN", "postgres://platform:secret@localhost:5433/platform?sslmode=disable"),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		KafkaBrokers:      strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		KafkaTopic:        getEnv("KAFKA_TOPIC", "transactions"),
		ClickHouseDSN:     getEnv("CLICKHOUSE_DSN", "clickhouse://localhost:9000?username=default&password="),
		JWTSecret:         getEnv("JWT_SECRET", "transaction-service-secret-key"),
		JWTExpiration:     time.Duration(getEnvAsInt("JWT_EXPIRATION", 15)) * time.Minute,
		RefreshExpiration: time.Duration(getEnvAsInt("REFRESH_EXPIRATION", 720)) * time.Hour,
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
