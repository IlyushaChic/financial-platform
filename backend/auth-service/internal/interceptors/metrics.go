package interceptors

import (
	"context"
	"time"

	"github.com/IlyushaChic/financial-platform/backend/shared/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// MetricsInterceptor собирает метрики Prometheus
func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start).Seconds()

		method := info.FullMethod
		statusCode := status.Code(err).String()

		metrics.GRPCRequestsTotal.WithLabelValues(method, statusCode).Inc()
		metrics.GRPCDuration.WithLabelValues(method).Observe(duration)

		return resp, err
	}
}
