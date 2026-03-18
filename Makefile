.PHONY: help setup dev-backend dev-frontend dev test build clean docker-up docker-down migrate

# Colors
BLUE=\033[0;34m
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m

help: ## Show this help message
	@echo "$(BLUE)Pintuotuo Development Commands$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

setup: ## Initialize development environment
	@echo "$(BLUE)Setting up development environment...$(NC)"
	@bash scripts/setup.sh

dev: dev-backend dev-frontend ## Run both backend and frontend

dev-backend: ## Run backend server (requires: docker-compose up -d)
	@echo "$(BLUE)Starting backend server...$(NC)"
	cd backend && go run main.go

dev-frontend: ## Run frontend dev server
	@echo "$(BLUE)Starting frontend dev server...$(NC)"
	cd frontend && npm run dev

test: ## Run all tests
	@echo "$(BLUE)Running tests...$(NC)"
	cd backend && go test ./... -v
	cd frontend && npm test

test-backend: ## Run backend tests only
	@echo "$(BLUE)Running backend tests...$(NC)"
	cd backend && export DATABASE_URL=postgresql://pintuotuo:dev_password_123@localhost:5433/pintuotuo_db?sslmode=disable; \
	export REDIS_URL=redis://localhost:6380; \
	export JWT_SECRET=pintuotuo-secret-key-dev; \
	export GIN_MODE=release; \
	go test -v -coverprofile=coverage.out \
		github.com/pintuotuo/backend \
		github.com/pintuotuo/backend/cache \
		github.com/pintuotuo/backend/config \
		github.com/pintuotuo/backend/errors \
		github.com/pintuotuo/backend/handlers \
		github.com/pintuotuo/backend/logger \
		github.com/pintuotuo/backend/metrics \
		github.com/pintuotuo/backend/middleware \
		github.com/pintuotuo/backend/models \
		github.com/pintuotuo/backend/routes \
		github.com/pintuotuo/backend/services/analytics \
		github.com/pintuotuo/backend/services/group \
		github.com/pintuotuo/backend/services/order \
		github.com/pintuotuo/backend/services/payment \
		github.com/pintuotuo/backend/services/product \
		github.com/pintuotuo/backend/services/token \
		github.com/pintuotuo/backend/services/user \
		github.com/pintuotuo/backend/tests \
		github.com/pintuotuo/backend/tests/integration
	cd backend && go tool cover -html=coverage.out

test-frontend: ## Run frontend tests only
	@echo "$(BLUE)Running frontend tests...$(NC)"
	cd frontend && npm test

lint: ## Run linters
	@echo "$(BLUE)Running linters...$(NC)"
	cd backend && go fmt ./... && go vet ./...
	cd frontend && npm run lint

format: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	cd backend && go fmt ./...
	cd frontend && npm run format

build: build-backend build-frontend ## Build both backend and frontend

build-backend: ## Build backend binary
	@echo "$(BLUE)Building backend...$(NC)"
	cd backend && go build -o ../bin/pintuotuo-backend

build-frontend: ## Build frontend for production
	@echo "$(BLUE)Building frontend...$(NC)"
	cd frontend && npm run build

docker-up: ## Start Docker containers
	@echo "$(BLUE)Starting Docker containers...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ Containers started$(NC)"
	@echo "  PostgreSQL: localhost:5432"
	@echo "  Redis: localhost:6379"
	@echo "  Kafka: localhost:29092"

docker-down: ## Stop Docker containers
	@echo "$(BLUE)Stopping Docker containers...$(NC)"
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

migrate: ## Run database migrations
	@echo "$(BLUE)Running migrations...$(NC)"
	cd backend && go run cmd/migrate/main.go

migrate-create: ## Create new migration file (usage: make migrate-create name=migration_name)
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=migration_name"; exit 1; fi
	@echo "Creating migration: $(name)"
	@touch "backend/migrations/$$(date +%03N)_$(name).sql"
	@echo "Created: backend/migrations/$$(date +%03N)_$(name).sql"

db-shell: ## Open PostgreSQL shell
	docker-compose exec postgres psql -U pintuotuo -d pintuotuo_db

db-reset: ## Reset database (WARNING: deletes all data)
	@read -p "$(YELLOW)This will delete all data. Continue? (y/n) $(NC)" -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose exec -T postgres psql -U pintuotuo -d pintuotuo_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"; \
		$(MAKE) migrate; \
		echo "$(GREEN)✓ Database reset and migrated$(NC)"; \
	fi

clean: ## Clean build artifacts and dependencies
	@echo "$(BLUE)Cleaning up...$(NC)"
	rm -rf bin/
	rm -rf backend/coverage.out
	rm -rf frontend/node_modules
	rm -rf frontend/dist
	@echo "$(GREEN)✓ Cleanup complete$(NC)"

install-hooks: ## Install git hooks
	@echo "$(BLUE)Installing git hooks...$(NC)"
	cp scripts/hooks/* .git/hooks/
	chmod +x .git/hooks/*
	@echo "$(GREEN)✓ Git hooks installed$(NC)"
