package tracer

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func TestTracerInit(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		ServiceName:  "test-service",
		CollectorURL: "host.docker.internal:4317",
		Insecure:     true,
		Timeout:      5 * time.Second,
	}
	tp, err := Init(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			t.Logf("failed to shutdown tracer: %v", err)
		}
	}()

	tracer := otel.Tracer("test-tracer")
	_, span := tracer.Start(ctx, "test-span")
	span.SetAttributes(attribute.String("test", "value"))
	span.End()

	time.Sleep(2 * time.Second)
	t.Log("Check Jaeger UI at http://localhost:16686 for service 'test-service'")
}
