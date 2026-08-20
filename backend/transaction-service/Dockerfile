FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/backend/transaction-service
RUN go mod download
RUN go build -o /transaction-service ./cmd

FROM alpine:latest
WORKDIR /app
COPY --from=builder /transaction-service .
EXPOSE 8080
CMD ["./transaction-service"]
