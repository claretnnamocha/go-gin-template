.DEFAULT_GOAL := help

# Variables
APP_NAME := go-gin-template
DOCKER_COMPOSE := docker-compose
DOCKER_COMPOSE_DEV := docker-compose -f docker-compose.yml -f docker-compose.dev.yml

up: ## Start all services
	@$(DOCKER_COMPOSE) up -d
	@echo "✓ Services started"
	@echo "✓ API: http://localhost:1954"
	@echo "✓ Swagger: http://localhost:1954/api/docs"
	@echo "✓ Health: http://localhost:1954/health"

down: ## Stop all services
	@$(DOCKER_COMPOSE) down
	@echo "✓ Services stopped"

restart: ## Restart API service
	@$(DOCKER_COMPOSE) restart api
	@echo "✓ API restarted"

build: ## Rebuild containers
	@$(DOCKER_COMPOSE) build --no-cache
	@echo "✓ Containers rebuilt"

ps: ## Show container status
	@$(DOCKER_COMPOSE) ps

logs: ## Follow API logs
	@$(DOCKER_COMPOSE) logs -f api

logs-all: ## Follow all logs
	@$(DOCKER_COMPOSE) logs -f

shell: ## Access API container
	@$(DOCKER_COMPOSE) exec api sh

db-shell: ## Access database
	@$(DOCKER_COMPOSE) exec postgres psql -U golang -d godb

redis-shell: ## Access Redis CLI
	@$(DOCKER_COMPOSE) exec redis redis-cli

dev: ## Start with hot reload
	@$(DOCKER_COMPOSE_DEV) up -d
	@echo "✓ Development mode started with hot reload"
	@echo "✓ API: http://localhost:1954"
	@echo "✓ Swagger: http://localhost:1954/api/docs"
	@$(DOCKER_COMPOSE_DEV) logs -f api

dev-build: ## Build and start with hot reload
	@$(DOCKER_COMPOSE_DEV) up -d --build
	@echo "✓ Development mode started with hot reload"
	@$(DOCKER_COMPOSE_DEV) logs -f api

dev-down: ## Stop development services
	@$(DOCKER_COMPOSE_DEV) down
	@echo "✓ Development services stopped"

clean: ## Remove all containers and volumes
	@$(DOCKER_COMPOSE) down -v --remove-orphans
	@$(DOCKER_COMPOSE_DEV) down -v --remove-orphans 2>/dev/null || true
	@echo "✓ Cleaned up containers and volumes"

migrate: ## Run migrations
	@$(DOCKER_COMPOSE) exec api migrate -path /app/migrations -database "postgresql://golang:secret@postgres:5432/godb?sslmode=disable" up
	@echo "✓ Migrations applied"

migrate-create: ## Create migration (name=create_users_table)
	@$(DOCKER_COMPOSE) exec api migrate create -ext sql -dir /app/migrations -seq $(name)
	@echo "✓ Migration created"

migrate-down: ## Rollback last migration
	@$(DOCKER_COMPOSE) exec api migrate -path /app/migrations -database "postgresql://golang:secret@postgres:5432/godb?sslmode=disable" down 1
	@echo "✓ Migration rolled back"

seed: ## Seed database
	@$(DOCKER_COMPOSE) exec api go run cmd/seed/main.go
	@echo "✓ Database seeded"

test: ## Run tests
	@$(DOCKER_COMPOSE) exec api go test ./...

test-unit: ## Run unit tests
	@$(DOCKER_COMPOSE) exec api go test ./tests/unit/...

test-e2e: ## Run E2E tests
	@$(DOCKER_COMPOSE) exec api go test ./tests/e2e/...

test-cov: ## Run tests with coverage
	@$(DOCKER_COMPOSE) exec api go test -coverprofile=coverage.out ./...
	@$(DOCKER_COMPOSE) exec api go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

lint: ## Run linter
	@$(DOCKER_COMPOSE) exec api golangci-lint run

format: ## Format code
	@$(DOCKER_COMPOSE) exec api gofmt -w .
	@$(DOCKER_COMPOSE) exec api goimports -w .
	@echo "✓ Code formatted"

swagger: ## Generate Swagger documentation
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api/main.go -o docs; \
	else \
		docker run --rm -v $(PWD):/app -w /app golang:1.23-alpine sh -c "go install github.com/swaggo/swag/cmd/swag@latest && swag init -g cmd/api/main.go -o docs"; \
	fi
	@echo "✓ Swagger docs generated"

health: ## Check service health
	@curl -sf http://localhost:1954/health | jq . || echo "Service unavailable"

tools: ## Install dev tools locally
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✓ Tools installed"

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
