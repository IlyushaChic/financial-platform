FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/backend/notification-service
RUN go mod download
RUN go build -o /notification-service ./cmd

FROM alpine:latest
WORKDIR /app
COPY --from=builder /notification-service .
EXPOSE 8080
CMD ["./notification-service"]
