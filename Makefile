.PHONY: help setup dev-backend dev-frontend dev dev-db test build clean docker-up docker-down migrate prompt-eval trace-verify

# Go 模块缓存：部分 IDE/沙箱会把 GOMODCACHE 指到不完整路径，导致 go test 报「源文件不存在」。
# 统一通过 scripts/ensure-go-modcache.sh 回退到 $HOME/go/pkg/mod（可被环境变量预先覆盖）。
export GOMODCACHE := $(shell "$(CURDIR)/scripts/ensure-go-modcache.sh")

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

prompt-eval: ## Promptfoo 回归（需环境变量 PROMPTFOO_BASE_URL、PROMPTFOO_API_KEY、PROMPTFOO_MODEL）
	@bash scripts/run_prompt_evals.sh

trace-verify: ## 在线验收 Tempo+OTel（支持 JWT/账号密码/短信登录自动验收用户链路）
	@bash scripts/verify_tracing_online.sh

test-backend: ## Run backend tests only
	@echo "$(BLUE)Running backend tests...$(NC)"
	cd backend && go test ./... -v -coverprofile=coverage.out
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

# 仅数据库依赖：看完整首页需后端 + 本目标（或全量 docker-up），只开 Vite 时接口会失败、内容为空。
dev-db: ## Start only Postgres + Redis (matches backend default DSN dev_password_123)
	@echo "$(BLUE)Starting postgres + redis only...$(NC)"
	DB_PASSWORD=dev_password_123 docker-compose up -d postgres redis
	@echo "$(GREEN)✓ Postgres: localhost:5432  Redis: localhost:6379$(NC)"
	@echo "  终端1: make dev-backend   终端2: make dev-frontend → http://localhost:5173/"

docker-down: ## Stop Docker containers
	@echo "$(BLUE)Stopping Docker containers...$(NC)"
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

migrate: ## Run database migrations
	@echo "$(BLUE)Running migrations...$(NC)"
	cd backend && go run cmd/migrate/main.go

reconcile-check: ## Full-database usage ledger check (api_usage_logs vs token_transactions); needs DATABASE_URL
	@echo "$(BLUE)Running usage reconciliation...$(NC)"
	cd backend && go run ./cmd/reconcile

capability-probe: ## Phase0: 用 DB 内 BYOK 密钥探测上游（GET /v1/models + 可选 POST embeddings）；需在部署机或能连库的环境执行，见 documentation/capability/README.md
	cd backend && go run ./cmd/capability-probe $(CAPABILITY_PROBE_FLAGS)

# 示例：make capability-probe CAPABILITY_PROBE_FLAGS='-out /tmp/cap.csv -provider openai -limit 5'
CAPABILITY_PROBE_FLAGS ?=

litellm-catalog-verify: ## 校验 litellm_proxy_config.yaml 覆盖库内 active SPU（需 DATABASE_URL；映射来自 model_providers）
	@echo "$(BLUE)Verifying LiteLLM model_list vs catalog...$(NC)"
	cd backend && go run ./cmd/litellm-catalog-sync -verify \
		-config ../deploy/litellm/litellm_proxy_config.yaml

litellm-catalog-verify-soft: ## 同上，-soft 仅警告（种子库与 P0 全量列表不一致时）
	cd backend && go run ./cmd/litellm-catalog-sync -verify -soft \
		-config ../deploy/litellm/litellm_proxy_config.yaml

litellm-catalog-generate: ## 由 DB 生成 model_list 片段 YAML（需 DATABASE_URL；可选 -map 见 deploy/litellm/README.md）
	@echo "$(BLUE)Generating LiteLLM model_list fragment from DB...$(NC)"
	cd backend && go run ./cmd/litellm-catalog-sync -generate \
		-out ../deploy/litellm/generated_model_list.fragment.yaml

# 可选覆盖：make litellm-catalog-assemble LITELLM_CATALOG_MAP=../path/to/map.json
LITELLM_CATALOG_MAP ?= ../deploy/litellm/provider_gateway_map.json

litellm-catalog-assemble: ## 由 DB 写出完整 litellm_proxy_config.yaml（显式 BYOK 列表；需 DATABASE_URL）
	@echo "$(BLUE)Assembling deploy/litellm/litellm_proxy_config.yaml from DB...$(NC)"
	cd backend && go run ./cmd/litellm-catalog-sync -write-full-proxy-config \
		../deploy/litellm/litellm_proxy_config.yaml \
		-map $(LITELLM_CATALOG_MAP)

probe-litellm: ## 读取 litellm_proxy_config.yaml 并探测网关 POST /v1/chat/completions（需 LITELLM_MASTER_KEY；可选 LITELLM_URL）
	@python3 scripts/probe_litellm_all_models.py --url "$${LITELLM_URL:-http://127.0.0.1:4000}"

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
