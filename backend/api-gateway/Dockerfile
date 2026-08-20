FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/backend/api-gateway
RUN go mod download
RUN go build -o /api-gateway ./cmd

FROM alpine:latest
WORKDIR /app
COPY --from=builder /api-gateway .
EXPOSE 8080
CMD ["./api-gateway"]
