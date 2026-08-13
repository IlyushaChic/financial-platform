package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	GRPCPort          string
	PostgresDSN       string
	RedisAddr         string
	RedisPassword     string
	JWTSecret         string
	JWTExpiration     time.Duration
	RefreshExpiration time.Duration
}

func Load() *Config {
	exp, _ := strconv.Atoi(getEnv("JWT_EXPIRATION", "15"))
	refreshExp, _ := strconv.Atoi(getEnv("REFRESH_EXPIRATION", "720"))
	return &Config{
		GRPCPort:          getEnv("GRPC_PORT", "50051"),
		PostgresDSN:       getEnv("POSTGRES_DSN", "postgres://platform:secret@localhost:5433/platform?sslmode=disable"),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		JWTSecret:         getEnv("JWT_SECRET", "supersecretkey"),
		JWTExpiration:     time.Duration(exp) * time.Minute,
		RefreshExpiration: time.Duration(refreshExp) * time.Hour,
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}
