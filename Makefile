.PHONY: help build test clean docker-up docker-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all services
	@echo "Building order-management-service..."
	@cd order-management-service && go build -o bin/main cmd/main.go
	@echo "Building order-reporting-service..."
	@cd order-reporting-service && go build -o bin/main cmd/main.go

test: ## Run tests for all modules
	@echo "Running tests..."
	@go test ./...

lint: ## Run golangci-lint on all modules
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

lint-shared: ## Run golangci-lint on shared module
	@echo "Running golangci-lint on shared module..."
	@cd shared && golangci-lint run

lint-command: ## Run golangci-lint on command service
	@echo "Running golangci-lint on command service..."
	@cd order-management-service && golangci-lint run

lint-query: ## Run golangci-lint on query service
	@echo "Running golangci-lint on query service..."
	@cd order-reporting-service && golangci-lint run

lint-fix: ## Run golangci-lint with --fix flag
	@echo "Running golangci-lint with auto-fix..."
	@golangci-lint run --fix ./...

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf order-management-service/bin
	@rm -rf order-reporting-service/bin

docker-up: ## Start all services with Docker Compose
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	@docker-compose down

docker-logs: ## Show logs from all services
	@docker-compose logs -f

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	@docker-compose build

dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	@go work init
	@go work use ./shared ./order-management-service ./order-reporting-service
	@go mod download

run-command-service: ## Run command service locally
	@echo "Running order-management-service..."
	@cd order-management-service && go run cmd/main.go

run-query-service: ## Run query service locally
	@echo "Running order-reporting-service..."
	@cd order-reporting-service && go run cmd/main.go

install-deps: ## Install dependencies
	@echo "Installing dependencies..."
	@go mod download
