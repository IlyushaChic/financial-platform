package interceptors

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// TracingInterceptor добавляет OpenTelemetry span для каждого gRPC запроса
func TracingInterceptor() grpc.UnaryServerInterceptor {
	tracer := otel.Tracer("auth-service")
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		span.SetAttributes(attribute.String("method", info.FullMethod))

		resp, err := handler(ctx, req)
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("error", err.Error()))
		}
		return resp, err
	}
}
