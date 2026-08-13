module github.com/IlyushaChic/financial-platform/backend/api-gateway

go 1.26

require (
	github.com/99designs/gqlgen v0.17.94
	github.com/IlyushaChic/financial-platform/backend/auth-service v0.0.0-00010101000000-000000000000
	github.com/IlyushaChic/financial-platform/backend/shared v0.0.0-00010101000000-000000000000
	github.com/IlyushaChic/financial-platform/backend/transaction-service v0.0.0-00010101000000-000000000000
	github.com/prometheus/client_golang v1.24.1
	github.com/vektah/gqlparser/v2 v2.5.36
	google.golang.org/grpc v1.83.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.21 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.70.1 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
)

require (
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/coder/websocket v1.8.15 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/rabbitmq/amqp091-go v1.13.0
	github.com/rs/zerolog v1.35.1
	github.com/sosodev/duration v1.4.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260803160001-6ac0973c030d // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

replace github.com/IlyushaChic/financial-platform/backend/auth-service => ../auth-service

replace github.com/IlyushaChic/financial-platform/backend/transaction-service => ../transaction-service

replace github.com/IlyushaChic/financial-platform/backend/shared => ../shared
