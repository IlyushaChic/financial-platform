module github.com/IlyushaChic/financial-platform/backend/notification-service

go 1.25.0

require (
	github.com/rabbitmq/amqp091-go v1.14.0
	github.com/rs/zerolog v1.35.1
)

require (
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

replace github.com/IlyushaChic/financial-platform/backend/shared => ../shared
