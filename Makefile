.PHONY: help build run stop clean test docker-build docker-up docker-down logs

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Go binaries
	@echo "Building API server..."
	cd backend && go build -o ../bin/api ./cmd/api
	@echo "Building worker service..."
	cd backend && go build -o ../bin/worker ./cmd/worker
	@echo "Build complete!"

run-api: ## Run the API server locally
	@echo "Starting API server..."
	cd backend && CONFIG_PATH=../config/config.yaml go run ./cmd/api

run-worker: ## Run the worker service locally
	@echo "Starting worker service..."
	cd backend && CONFIG_PATH=../config/config.yaml go run ./cmd/worker

test: ## Run tests
	cd backend && go test -v ./...

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	cd docker && docker-compose build

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	cd docker && docker-compose up -d
	@echo "Services started!"
	@echo "API: http://localhost:8080"
	@echo "RabbitMQ UI: http://localhost:15672"

docker-down: ## Stop all services
	@echo "Stopping services..."
	cd docker && docker-compose down

docker-clean: ## Stop services and remove volumes
	@echo "Cleaning up..."
	cd docker && docker-compose down -v

logs: ## View Docker logs
	cd docker && docker-compose logs -f

logs-api: ## View API logs
	cd docker && docker-compose logs -f api

logs-worker: ## View worker logs
	cd docker && docker-compose logs -f worker

db-migrate: ## Run database migrations
	@echo "Running migrations..."
	docker exec -i pushlab-postgres psql -U pushlab -d pushlab < migrations/001_initial_schema.sql
	@echo "Migrations complete!"

db-shell: ## Open PostgreSQL shell
	docker exec -it pushlab-postgres psql -U pushlab -d pushlab

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf backend/api backend/worker

deps: ## Install Go dependencies
	cd backend && go mod download
	cd backend && go mod tidy

format: ## Format Go code
	cd backend && go fmt ./...

lint: ## Run linters
	cd backend && go vet ./...

setup: ## Initial setup (copy config files)
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file - please edit it with your values"; \
	fi
	@if [ ! -f config/config.yaml ]; then \
		cp config/config.yaml config/config.yaml; \
		echo "Created config/config.yaml"; \
	fi

dev: setup ## Setup and start development environment
	@echo "Starting development environment..."
	make docker-up
	@echo ""
	@echo "Development environment ready!"
	@echo "API: http://localhost:8080"
	@echo "Health check: curl http://localhost:8080/health"
