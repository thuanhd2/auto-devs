# Load environment variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Default values if not set in environment
DB_HOST ?= 127.0.0.1
DB_PORT ?= 5432
DB_USERNAME ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= autodevs_dev
MIGRATIONS_PATH ?= ./migrations

# Database URL for migrations
DATABASE_URL = postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: help
help: ## Show this help message
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

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
	@go test ./...

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@echo "Clean completed"

.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"