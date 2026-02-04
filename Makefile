# ============================================
# PPOB.ID Backend - Makefile
# ============================================

.PHONY: help build run test clean docker-up docker-down docker-logs migrate-up migrate-down

# Default target
.DEFAULT_GOAL := help

# ============================================
# HELP
# ============================================
help: ## Show this help message
	@echo "PPOB.ID Backend - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

# ============================================
# DEVELOPMENT
# ============================================
install: ## Install dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

build: ## Build the application
	@echo "Building application..."
	go build -o bin/ppob-api cmd/api/main.go

run: ## Run the application
	@echo "Running application..."
	go run cmd/api/main.go

test: ## Run tests
	@echo "Running tests..."
	go test -v -cover ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

# ============================================
# DOCKER
# ============================================
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker-compose build

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	docker-compose up -d
	@echo "Services started! API: http://localhost:8080"

docker-down: ## Stop all services
	@echo "Stopping services..."
	docker-compose down

docker-restart: ## Restart all services
	@echo "Restarting services..."
	docker-compose restart

docker-logs: ## Show logs
	@echo "Showing logs..."
	docker-compose logs -f

docker-logs-api: ## Show API logs only
	@echo "Showing API logs..."
	docker-compose logs -f ppob_backend

docker-ps: ## Show running containers
	@echo "Running containers:"
	docker-compose ps

docker-clean: ## Remove containers, volumes, and images
	@echo "Cleaning Docker resources..."
	docker-compose down -v --rmi all

# ============================================
# DOCKER PRODUCTION
# ============================================
docker-prod-build: ## Build production Docker image
	@echo "Building production Docker image..."
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml build

docker-prod-up: ## Start production services
	@echo "Starting production services..."
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
	@echo "Production services started!"

docker-prod-down: ## Stop production services
	@echo "Stopping production services..."
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml down

docker-prod-logs: ## Show production logs
	@echo "Showing production logs..."
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml logs -f

# ============================================
# DATABASE MIGRATIONS
# ============================================
migrate-create: ## Create a new migration (usage: make migrate-create name=create_users_table)
	@if [ -z "$(name)" ]; then \
		echo "Error: Migration name is required. Usage: make migrate-create name=create_users_table"; \
		exit 1; \
	fi
	@echo "Creating migration: $(name)"
	@timestamp=$$(date +%Y%m%d%H%M%S); \
	filename="migrations/$${timestamp}_$(name).sql"; \
	touch $$filename; \
	echo "-- Migration: $(name)" > $$filename; \
	echo "-- Created at: $$(date)" >> $$filename; \
	echo "" >> $$filename; \
	echo "-- Up Migration" >> $$filename; \
	echo "" >> $$filename; \
	echo "" >> $$filename; \
	echo "-- Down Migration" >> $$filename; \
	echo "-- Run manually if needed" >> $$filename; \
	echo "Migration created: $$filename"

migrate-status: ## Check migration status (manual - check migrations/ folder)
	@echo "Checking migrations..."
	@ls -la migrations/

# ============================================
# ENVIRONMENT
# ============================================
env-example: ## Create .env from .env.example
	@if [ ! -f .env ]; then \
		echo "Creating .env from .env.example..."; \
		cp .env.example .env; \
		echo ".env created! Please update with your credentials."; \
	else \
		echo ".env already exists. Skipping..."; \
	fi

env-check: ## Validate environment variables
	@echo "Checking environment variables..."
	@go run cmd/api/main.go --check-env 2>&1 | head -1 || echo "Config loaded successfully!"

# ============================================
# QUALITY CHECKS
# ============================================
check: fmt lint test ## Run all quality checks

ci: check ## Run CI checks (fmt, lint, test)
	@echo "CI checks passed!"

# ============================================
# QUICK START
# ============================================
setup: env-example install ## Setup project (copy .env.example, install deps)
	@echo "Setup complete! Next steps:"
	@echo "  1. Update .env with your credentials"
	@echo "  2. Run: make docker-up"
	@echo "  3. Visit: http://localhost:8080/health"

dev: docker-up docker-logs ## Start development environment and show logs

# ============================================
# UTILITIES
# ============================================
health: ## Check API health
	@echo "Checking API health..."
	@curl -s http://localhost:8080/health | jq . || echo "API not running or jq not installed"

ps: docker-ps ## Show running containers

logs: docker-logs ## Show all logs

# ============================================
# DATABASE UTILITIES
# ============================================
db-shell: ## Connect to PostgreSQL database
	@echo "Connecting to database..."
	docker-compose exec ppob_postgres psql -U ppob_user -d ppob_db

db-backup: ## Backup database
	@echo "Backing up database..."
	@timestamp=$$(date +%Y%m%d_%H%M%S); \
	docker-compose exec -T ppob_postgres pg_dump -U ppob_user ppob_db > backup_$${timestamp}.sql; \
	echo "Backup created: backup_$${timestamp}.sql"

db-restore: ## Restore database (usage: make db-restore file=backup_20240101_120000.sql)
	@if [ -z "$(file)" ]; then \
		echo "Error: Backup file is required. Usage: make db-restore file=backup_20240101_120000.sql"; \
		exit 1; \
	fi
	@echo "Restoring database from $(file)..."
	docker-compose exec -T ppob_postgres psql -U ppob_user -d ppob_db < $(file)
	@echo "Database restored!"

redis-cli: ## Connect to Redis CLI
	@echo "Connecting to Redis..."
	docker-compose exec ppob_redis redis-cli

redis-flush: ## Flush Redis cache (WARNING: deletes all cache)
	@echo "Flushing Redis cache..."
	docker-compose exec ppob_redis redis-cli FLUSHALL
	@echo "Redis cache flushed!"

# ============================================
# MONITORING
# ============================================
stats: ## Show Docker stats
	@echo "Container stats:"
	docker stats --no-stream

top: ## Show container processes
	@echo "Container processes:"
	docker-compose top
