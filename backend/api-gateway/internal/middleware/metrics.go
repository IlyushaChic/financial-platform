package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/IlyushaChic/financial-platform/backend/shared/metrics"
)

// MetricsMiddleware собирает Prometheus метрики
func MetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r)
			duration := time.Since(start).Seconds()

			metrics.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(ww.status)).Inc()
			metrics.HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
		})
	}
}
