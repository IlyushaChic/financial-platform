BIN_DIR := $(CURDIR)/bin
SERVICES := api-gateway auth-service transaction-service notification-service
BINS := $(addprefix $(BIN_DIR)/, $(SERVICES))

DB_URL ?= postgres://platform:secret@host.docker.internal:5433/platform?sslmode=disable
CLICKHOUSE_URL ?= clickhouse://host.docker.internal:9000?username=default&password=

.PHONY: all build clean lint test run stop help migrate-up migrate-down

all: build

build: $(BINS)

$(BIN_DIR)/%:
	@echo "Building $*..."
	@mkdir -p $(BIN_DIR)
	@cd $* && go build -o $(BIN_DIR)/$* ./cmd/

clean:
	@rm -rf $(BIN_DIR)
	@echo "Cleaned binaries"

lint:
	@echo "Running linters for each module..."
	@for dir in api-gateway auth-service transaction-service notification-service shared; do \
		echo "Linting $$dir..."; \
		(cd $$dir && golangci-lint run ./...) || exit 1; \
	done

test:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@echo "Coverage report saved to coverage.out"
	@go tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage report generated: coverage.html"


run: build
	@for bin in $(BINS); do \
		echo "Starting $$(basename $$bin)..." ; \
		$$bin & \
	done
	@echo "All services started in background."

stop:
	@pkill -f "bin/(api-gateway|auth-service|transaction-service|notification-service)" || true
	@echo "Stopped all services"


migrate-up:
	@echo "Running PostgreSQL migrations up..."
	@docker run --rm -v $(CURDIR)/transaction-service/migrations/postgres:/migrations \
		migrate/migrate \
		-path=/migrations -database "$(DB_URL)" up
	@echo "Running ClickHouse migrations up..."
	@docker run --rm -v $(CURDIR)/transaction-service/migrations/clickhouse:/migrations \
		migrate/migrate \
		-path=/migrations -database "$(CLICKHOUSE_URL)" up

migrate-down:
	@echo "Rolling back PostgreSQL migrations..."
	@docker run --rm -v $(CURDIR)/transaction-service/migrations/postgres:/migrations \
		migrate/migrate \
		-path=/migrations -database "$(DB_URL)" down
	@echo "Rolling back ClickHouse migrations..."
	@docker run --rm -v $(CURDIR)/transaction-service/migrations/clickhouse:/migrations \
		migrate/migrate \
		-path=/migrations -database "$(CLICKHOUSE_URL)" down


help:
	@echo "Available targets:"
	@echo "  build   - Build all services"
	@echo "  clean   - Remove binaries"
	@echo "  lint    - Run golangci-lint on each module"
	@echo "  test    - Run tests with coverage (HTML report)"
	@echo "  run     - Build and run all services in background"
	@echo "  stop    - Stop all running services"
	@echo "  migrate-up   - Apply database migrations (via Docker)"
	@echo "  migrate-down - Rollback database migrations (via Docker)"