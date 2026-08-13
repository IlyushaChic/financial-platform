package main

import (
	"context"
	"net/http"
	"time"

	"github.com/IlyushaChic/financial-platform/backend/shared/logger"
	"github.com/IlyushaChic/financial-platform/backend/shared/metrics"
	"github.com/IlyushaChic/financial-platform/backend/shared/tracer"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logCfg := logger.Config{
		Level: "debug",
		JSON:  true,
	}
	logger := logger.New(logCfg)

	ctx := context.Background()
	tracerCfg := tracer.Config{
		ServiceName:  "api-gateway",
		CollectorURL: "jaeger:4317", //  host.docker.internal:4317 для локальной разработки
		Insecure:     true,
		Timeout:      5 * time.Second,
	}
	tp, err := tracer.Init(ctx, tracerCfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to init tracer")
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error().Err(err).Msg("failed to shutdown tracer")
		}
	}()

	logger.Info().Msg("Service started with logger, tracer and metrics")

	metrics.IncHTTPRequest("GET", "/health", "200")

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		logger.Info().Msg("Starting metrics server on :9091")
		server := &http.Server{
			Addr:         ":9091",
			Handler:      nil,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}
		if err := server.ListenAndServe(); err != nil {
			logger.Error().Err(err).Msg("Metrics server failed")
		}
	}()

	// Чтобы сервер не завершился сразу, добавь блокировку (например, select{})
	select {}

	// Далее запускаем HTTP сервер, gRPC сервер и т.д.
}
