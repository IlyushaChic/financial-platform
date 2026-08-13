package main

import (
	"context"
	"database/sql"
	"net"
	"time"

	_ "github.com/lib/pq"
	redisdb "github.com/redis/go-redis/v9"

	"github.com/IlyushaChic/financial-platform/backend/auth-service/internal/application"
	"github.com/IlyushaChic/financial-platform/backend/auth-service/internal/config"
	redisrepo "github.com/IlyushaChic/financial-platform/backend/auth-service/internal/infrastructure/cache/redis" // алиас для твоего пакета
	"github.com/IlyushaChic/financial-platform/backend/auth-service/internal/infrastructure/jwt"
	postgresrepo "github.com/IlyushaChic/financial-platform/backend/auth-service/internal/infrastructure/persistence/postgres"
	"github.com/IlyushaChic/financial-platform/backend/auth-service/internal/interceptors"
	proto "github.com/IlyushaChic/financial-platform/backend/auth-service/proto"
	"github.com/IlyushaChic/financial-platform/backend/shared/logger"
	"github.com/IlyushaChic/financial-platform/backend/shared/tracer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()

	// Logger
	logCfg := logger.Config{Level: "debug", JSON: true}
	log := logger.New(logCfg)

	// Tracer
	ctx := context.Background()
	tp, err := tracer.Init(ctx, tracer.Config{
		ServiceName:  "auth-service",
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

	// Redis – используем redisdb.NewClient
	rdb := redisdb.NewClient(&redisdb.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}

	// Repositories
	userRepo := postgresrepo.NewUserRepo(db)
	sessionRepo := redisrepo.NewSessionRepo(rdb) // используем алиас redisrepo
	jwtMgr := jwt.NewManager(cfg.JWTSecret, cfg.JWTExpiration, cfg.RefreshExpiration)

	// gRPC сервер с middleware
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.LoggingInterceptor(log),
			interceptors.TracingInterceptor(),
			interceptors.MetricsInterceptor(),
		),
	)

	authSvc := application.NewAuthService(userRepo, sessionRepo, jwtMgr)
	proto.RegisterAuthServiceServer(grpcServer, authSvc)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
	log.Info().Msgf("Auth service listening on port %s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to serve")
	}
}
