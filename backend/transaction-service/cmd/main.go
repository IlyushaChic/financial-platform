package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	redisdb "github.com/redis/go-redis/v9"

	"github.com/IlyushaChic/financial-platform/backend/shared/logger"
	"github.com/IlyushaChic/financial-platform/backend/shared/tracer"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/application/command"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/config"
	redisrepo "github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/infrastructure/cache/redis"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/infrastructure/messaging/kafka"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/infrastructure/messaging/rabbitmq"
	clickhouserepo "github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/infrastructure/persistence/clickhouse"
	postgresrepo "github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/infrastructure/persistence/postgres"
)

func main() {
	cfg := config.Load()

	// ---------- Logger ----------
	logCfg := logger.Config{Level: "debug", JSON: true}
	log := logger.New(logCfg)

	// ---------- Tracer ----------
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

	// ---------- PostgreSQL ----------
	db, err := sql.Open("postgres", cfg.PostgresDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer db.Close()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)

	// ---------- Redis ----------
	rdb := redisdb.NewClient(&redisdb.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer rdb.Close()

	// ---------- ClickHouse ----------
	clickhouseRepo, err := clickhouserepo.NewAnalyticsRepo(cfg.ClickHouseDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to clickhouse")
	}
	defer clickhouseRepo.Close()

	// ---------- Kafka Producer ----------
	kafkaPublisher, err := kafka.NewKafkaPublisher(cfg.KafkaBrokers, cfg.KafkaTopic)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kafka publisher")
	}
	defer kafkaPublisher.Close()

	// ---------- RabbitMQ Producer ----------
	rabbitConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to rabbitmq")
	}
	defer rabbitConn.Close()

	notificationProducer, err := rabbitmq.NewProducer(rabbitConn)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create rabbitmq producer")
	}
	defer notificationProducer.Close()

	// ---------- Repositories ----------
	accountRepo := postgresrepo.NewAccountRepo(db)
	transactionRepo := postgresrepo.NewTransactionRepo(db)

	// ---------- Distributed Lock ----------
	locker := redisrepo.NewDistributedLocker(rdb)

	// ---------- Kafka Consumer ----------
	consumer, err := kafka.NewEventConsumer(
		cfg.KafkaBrokers,
		"transaction-consumer-group",
		cfg.KafkaTopic,
		cfg.KafkaTopic+"_dlq",
		clickhouseRepo,
		rdb,
		log,
		kafkaPublisher,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kafka consumer")
	}
	defer consumer.Close()

	go func() {
		log.Info().Msg("starting kafka consumer...")
		if err := consumer.Start(ctx); err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("kafka consumer stopped with error")
		}
	}()

	// ---------- Transfer Handler ----------
	transferHandler := command.NewTransferHandler(
		accountRepo,
		transactionRepo,
		locker,
		db,
		kafkaPublisher,
		notificationProducer,
	)

	// Используем transferHandler, чтобы избежать ошибки компиляции
	// В реальном проекте здесь будет вызов через HTTP/gRPC
	_ = transferHandler

	log.Info().Msg("transaction service started successfully")

	// ---------- Graceful Shutdown ----------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down gracefully...")
}
