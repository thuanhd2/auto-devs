# Load environment variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Default values if not set in environment
AUTODEVS_DB_HOST ?= 127.0.0.1
AUTODEVS_DB_PORT ?= 5432
AUTODEVS_DB_USERNAME ?= postgres
AUTODEVS_DB_PASSWORD ?= postgres
AUTODEVS_DB_NAME ?= autodevs_dev
MIGRATIONS_PATH ?= ./migrations

# Database URL for migrations
DATABASE_URL = postgres://$(AUTODEVS_DB_USERNAME):$(AUTODEVS_DB_PASSWORD)@$(AUTODEVS_DB_HOST):$(AUTODEVS_DB_PORT)/$(AUTODEVS_DB_NAME)?sslmode=disable

.PHONY: help
help: ## Show this help message
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: swagger
swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/server/main.go
	@echo "Swagger documentation generated successfully!"

.PHONY: swagger-serve
swagger-serve: ## Serve Swagger UI (requires server to be running)
	@echo "Swagger UI available at:"
	@echo "  http://localhost:8098/swagger/index.html"
	@echo "  http://localhost:8098/ (redirects to Swagger UI)"

.PHONY: migrate-up
migrate-up: ## Run all pending migrations
	@echo "Running database migrations..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up
	@echo "Migrations completed successfully"

.PHONY: migrate-down
migrate-down: ## Rollback the last migration
	@echo "Rolling back last migration..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down 1
	@echo "Rollback completed successfully"

.PHONY: migrate-reset
migrate-reset: ## Rollback all migrations
	@echo "Rolling back all migrations..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down
	@echo "All migrations rolled back"

.PHONY: migrate-force
migrate-force: ## Force migration to specific version (usage: make migrate-force VERSION=1)
	@if [ -z "$(VERSION)" ]; then echo "Usage: make migrate-force VERSION=<version>"; exit 1; fi
	@echo "Forcing migration to version $(VERSION)..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" force $(VERSION)
	@echo "Migration forced to version $(VERSION)"

.PHONY: migrate-version
migrate-version: ## Show current migration version
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" version

.PHONY: migrate-create
migrate-create: ## Create a new migration (usage: make migrate-create name=migration_name)
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=<migration_name>"; exit 1; fi
	@echo "Creating new migration: $(name)"
	@migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)
	@echo "Migration files created"

.PHONY: db-setup
db-setup: ## Setup database (run migrations)
	@echo "Setting up database..."
	@make migrate-up
	@echo "Database setup completed"

.PHONY: build
build: ## Build the application
	@echo "Building application..."
	@go build -o bin/autodevs cmd/server/main.go
	@echo "Build completed"

.PHONY: run
run: ## Run the application
	@echo "Starting application..."
	@go run cmd/server/main.go

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@go test ./... -v

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@echo "Clean completed"

.PHONY: clean-tools
clean-tools: ## Clean external tools

.PHONY: build-worker
build-worker: ## Build the job worker binary
	@echo "Building job worker..."
	@go build -o bin/worker cmd/worker/main.go
	@echo "Worker build completed"

.PHONY: run-worker
run-worker: ## Run the job worker (requires Redis)
	@echo "Starting job worker..."
	@./scripts/run-worker.sh

.PHONY: run-worker-verbose
run-worker-verbose: ## Run the job worker with verbose logging
	@echo "Starting job worker with verbose logging..."
	@./scripts/run-worker.sh -v

.PHONY: run-worker-named
run-worker-named: ## Run the job worker with custom name (usage: make run-worker-named name=worker-1)
	@if [ -z "$(name)" ]; then echo "Usage: make run-worker-named name=<worker_name>"; exit 1; fi
	@echo "Starting job worker: $(name)"
	@./scripts/run-worker.sh -n $(name)

.PHONY: worker-help
worker-help: ## Show worker help
	@./scripts/run-worker.sh -h
	@echo "Cleaning external tools..."
	@rm -rf external-tools/
	@echo "External tools cleaned!"

.PHONY: wire
wire: ## Generate Wire dependency injection code
	@echo "Generating Wire dependency injection code..."
	@go generate ./internal/di
	@echo "Wire code generated successfully!"

.PHONY: build-tools
build-tools: ## Download external tools (mockery)
	@echo "Downloading external tools..."
	@mkdir -p external-tools
	@if [ ! -f external-tools/mockery ]; then \
		echo "Downloading mockery v3.2.4..."; \
		curl -L -o external-tools/mockery.tar.gz https://github.com/vektra/mockery/releases/download/v3.2.4/mockery_3.2.4_$(shell uname -s)_$(shell uname -m).tar.gz; \
		tar -xzf external-tools/mockery.tar.gz -C external-tools; \
		rm external-tools/mockery.tar.gz; \
		chmod +x external-tools/mockery; \
		echo "Mockery downloaded successfully!"; \
	else \
		echo "Mockery already exists in external-tools/"; \
	fi

.PHONY: mocks
mocks: build-tools ## Generate mocks using mockery from external-tools
	@echo "Generating mocks..."
	@./external-tools/mockery --config .mockery.yaml
	@echo "Mocks generated successfully!"

.PHONY: mocks-clean
mocks-clean: ## Clean generated mocks
	@echo "Cleaning generated mocks..."
	@rm -rf internal/mocks
	@mkdir -p internal/mocks/usecase internal/mocks/repository
	@echo "Mocks cleaned!"

.PHONY: mocks-regen
mocks-regen: mocks-clean mocks ## Regenerate all mocks
	@echo "All mocks regenerated successfully!"

.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"