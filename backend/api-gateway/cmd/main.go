package main

import (
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/IlyushaChic/financial-platform/backend/api-gateway/internal/clients"
	"github.com/IlyushaChic/financial-platform/backend/api-gateway/internal/graphql/generated"
	"github.com/IlyushaChic/financial-platform/backend/api-gateway/internal/graphql/resolver"
	"github.com/IlyushaChic/financial-platform/backend/api-gateway/internal/middleware"
	"github.com/IlyushaChic/financial-platform/backend/shared/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Логгер
	logCfg := logger.Config{Level: "debug", JSON: true}
	log := logger.New(logCfg)

	// gRPC клиенты
	authClient, err := clients.NewAuthClient("localhost:50051")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create auth client")
	}
	defer authClient.Close()

	transactionClient, err := clients.NewTransactionClient("localhost:50052")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create transaction client")
	}
	defer transactionClient.Close()

	// Resolver
	res := resolver.NewResolver(authClient, transactionClient)

	// GraphQL handler
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: res}))

	// Применяем middleware к /query
	queryHandler := middleware.AuthMiddleware(authClient)(
		middleware.LoggingMiddleware(log)(
			middleware.MetricsMiddleware()(srv),
		),
	)

	// Маршруты
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", queryHandler)
	http.Handle("/metrics", promhttp.Handler()) // эндпоинт для метрик
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Запуск
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Info().Str("port", port).Msg("API Gateway started")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
