package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTPRequestsTotal – общее количество HTTP запросов
var HTTPRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "endpoint", "status"},
)

// HTTPRequestDuration – гистограмма длительности запросов
var HTTPRequestDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "endpoint"},
)

// ActiveRequests – количество одновременно обрабатываемых запросов
var ActiveRequests = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "http_active_requests",
		Help: "Number of active HTTP requests",
	},
)

// DatabaseQueriesTotal – счётчик запросов к БД
var DatabaseQueriesTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "database_queries_total",
		Help: "Total number of database queries",
	},
	[]string{"driver", "operation"},
)

// DatabaseQueryDuration – гистограмма длительности запросов к БД
var DatabaseQueryDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "database_query_duration_seconds",
		Help:    "Database query duration in seconds",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	},
	[]string{"driver", "operation"},
)

// KafkaProducerMessagesTotal – счётчик отправленных сообщений в Kafka
var KafkaProducerMessagesTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_producer_messages_total",
		Help: "Total number of Kafka messages produced",
	},
	[]string{"topic"},
)

// RabbitMQPublishedTotal – счётчик опубликованных сообщений в RabbitMQ
var RabbitMQPublishedTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "rabbitmq_published_total",
		Help: "Total number of RabbitMQ messages published",
	},
	[]string{"exchange", "routing_key"},
)

// IncHTTPRequest увеличивает счётчик запросов
func IncHTTPRequest(method, endpoint, status string) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
}

// ObserveHTTPDuration записывает время выполнения запроса
func ObserveHTTPDuration(method, endpoint string, seconds float64) {
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(seconds)
}

// IncDatabaseQuery увеличивает счётчик запросов к БД
func IncDatabaseQuery(driver, operation string) {
	DatabaseQueriesTotal.WithLabelValues(driver, operation).Inc()
}

// ObserveDatabaseQueryDuration записывает время выполнения запроса к БД
func ObserveDatabaseQueryDuration(driver, operation string, seconds float64) {
	DatabaseQueryDuration.WithLabelValues(driver, operation).Observe(seconds)
}

// GRPCRequestsTotal – общее количество gRPC запросов
var GRPCRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "grpc_requests_total",
		Help: "Total number of gRPC requests",
	},
	[]string{"method", "status"},
)

// GRPCDuration – гистограмма длительности gRPC запросов
var GRPCDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "grpc_request_duration_seconds",
		Help:    "gRPC request duration in seconds",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method"},
)
