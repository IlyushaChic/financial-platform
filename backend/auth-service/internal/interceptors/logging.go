package interceptors

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor логирует входящие gRPC запросы и ответы
func LoggingInterceptor(logger *zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		statusCode := status.Code(err).String()

		event := logger.Info()
		if err != nil {
			event = logger.Error().Err(err)
		}

		event.Str("method", info.FullMethod).
			Str("status", statusCode).
			Dur("duration", duration).
			Msg("gRPC request")

		return resp, err
	}
}
