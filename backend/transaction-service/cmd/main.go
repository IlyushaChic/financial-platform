package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	redisdb "github.com/redis/go-redis/v9"

	"github.com/IlyushaChic/financial-platform/backend/shared/logger"
	"github.com/IlyushaChic/financial-platform/backend/shared/tracer"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/application/command"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/config"
	redisrepo "github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/infrastructure/cache/redis"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/infrastructure/messaging/kafka"
	clickhouserepo "github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/infrastructure/persistence/clickhouse"
	postgresrepo "github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/infrastructure/persistence/postgres"
)

func main() {
	cfg := config.Load()

	// Logger
	logCfg := logger.Config{Level: "debug", JSON: true}
	log := logger.New(logCfg)

	// Tracer
	ctx := context.Background()
	tp, err := tracer.Init(ctx, tracer.Config{
		ServiceName:  "transaction-service",
		CollectorURL: "jaeger:4317",
		Insecure:     true,
		Timeout:      5 * time.Second,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init tracer")
	}
	defer tp.Shutdown(ctx)

	// PostgreSQL
	db, err := sql.Open("postgres", cfg.PostgresDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer db.Close()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)

	// Redis
	rdb := redisdb.NewClient(&redisdb.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}

	// ClickHouse
	clickhouseRepo, err := clickhouserepo.NewAnalyticsRepo(cfg.ClickHouseDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to clickhouse")
	}
	defer clickhouseRepo.Close()

	// Репозитории
	accountRepo := postgresrepo.NewAccountRepo(db)
	transactionRepo := postgresrepo.NewTransactionRepo(db)

	// Distributed Lock
	locker := redisrepo.NewDistributedLocker(rdb)

	// Kafka Producer
	kafkaPublisher, err := kafka.NewKafkaPublisher(cfg.KafkaBrokers, cfg.KafkaTopic)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kafka publisher")
	}
	defer kafkaPublisher.Close()

	// Use cases
	transferHandler := command.NewTransferHandler(accountRepo, transactionRepo, locker, db, kafkaPublisher)
	_ = transferHandler // временно

	// ---------- ЗАПУСК КОНСЬЮМЕРА ----------
	// В main.go после создания продюсера:
	dlqTopic := cfg.KafkaTopic + "_dlq" // "transactions_dlq"

	consumer, err := kafka.NewEventConsumer(
		cfg.KafkaBrokers,
		"transaction-consumer-group",
		cfg.KafkaTopic,
		dlqTopic,
		clickhouseRepo,
		rdb,
		log,
		kafkaPublisher, // передаём продюсера для отправки в DLQ
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kafka consumer")
	}
	defer consumer.Close()

	// Запускаем консьюмера в горутине
	go func() {
		log.Info().Msg("starting kafka consumer...")
		if err := consumer.Start(ctx); err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("kafka consumer stopped with error")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down gracefully...")
}
