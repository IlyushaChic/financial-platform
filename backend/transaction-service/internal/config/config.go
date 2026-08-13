package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	GRPCPort          string
	PostgresDSN       string
	RedisAddr         string
	RedisPassword     string
	KafkaBrokers      []string
	KafkaTopic        string
	ClickHouseDSN     string
	RabbitMQURL       string // <-- добавлено
	JWTSecret         string
	JWTExpiration     time.Duration
	RefreshExpiration time.Duration
}

func Load() *Config {
	return &Config{
		GRPCPort:          getEnv("GRPC_PORT", "50052"),
		PostgresDSN:       getEnv("POSTGRES_DSN", "postgres://platform:secret@localhost:5433/platform?sslmode=disable"),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		KafkaBrokers:      strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		KafkaTopic:        getEnv("KAFKA_TOPIC", "transactions"),
		ClickHouseDSN:     getEnv("CLICKHOUSE_DSN", "clickhouse://localhost:9000?username=default&password="),
		RabbitMQURL:       getEnv("RABBITMQ_URL", "amqp://admin:admin@localhost:5672/"),
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
