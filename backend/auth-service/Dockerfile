FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/backend/auth-service
RUN go mod download
RUN go build -o /auth-service ./cmd

FROM alpine:latest
WORKDIR /app
COPY --from=builder /auth-service .
EXPOSE 8080
CMD ["./auth-service"]
