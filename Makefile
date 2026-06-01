# ============================================================
# Pote - Monorepo Makefile
# ============================================================

.PHONY: help dev-infra dev-infra-down frontend \
        auth-service user-service chat-service message-service \
        realtime-service ai-orchestrator notification-service \
        secrets-service audit-service \
        lint test build clean

# ---------- Help ----------
help: ## Show this help message
	@echo.
	@echo   Pote - Available Commands
	@echo   =========================
	@echo.
	@echo   Infrastructure:
	@echo     make dev-infra         - Start local Postgres + Redis
	@echo     make dev-infra-down    - Stop local infrastructure
	@echo.
	@echo   Frontend:
	@echo     make frontend          - Start Next.js dev server
	@echo.
	@echo   Services:
	@echo     make auth-service      - Start auth service
	@echo     make user-service      - Start user service
	@echo     make chat-service      - Start chat service
	@echo     make message-service   - Start message service
	@echo     make realtime-service  - Start realtime service
	@echo     make ai-orchestrator   - Start AI orchestrator
	@echo     make notification-service - Start notification service
	@echo     make secrets-service   - Start secrets service
	@echo     make audit-service     - Start audit service
	@echo.
	@echo   Utilities:
	@echo     make lint              - Run linters across all services
	@echo     make test              - Run tests across all services
	@echo     make build             - Build all services
	@echo     make clean             - Clean build artifacts
	@echo.

# ---------- Infrastructure ----------
dev-infra: ## Start local development infrastructure (Postgres + Redis)
	docker compose up -d

dev-infra-down: ## Stop local development infrastructure
	docker compose down

# ---------- Frontend ----------
frontend: ## Start Next.js development server
	cd frontend && pnpm dev

# ---------- Go Services ----------
auth-service: ## Start auth service
	cd services/auth-service && go run cmd/server/main.go

user-service: ## Start user service
	cd services/user-service && go run cmd/server/main.go

chat-service: ## Start chat service
	cd services/chat-service && go run cmd/server/main.go

message-service: ## Start message service
	cd services/message-service && go run cmd/server/main.go

realtime-service: ## Start realtime service
	cd services/realtime-service && go run cmd/server/main.go

notification-service: ## Start notification service
	cd services/notification-service && go run cmd/server/main.go

secrets-service: ## Start secrets service
	cd services/secrets-service && go run cmd/server/main.go

audit-service: ## Start audit service
	cd services/audit-service && go run cmd/server/main.go

# ---------- Python Services ----------
ai-orchestrator: ## Start AI orchestrator service
	cd services/ai-orchestrator-service && uvicorn app.main:app --reload --port 8086

# ---------- Utilities ----------
lint: ## Run linters across all Go services
	@echo Running Go linters...
	@for /D %%d in (services\*-service) do ( \
		echo Linting %%d... && \
		cd %%d && golangci-lint run ./... && cd ..\.. \
	)

test: ## Run tests across all Go services
	@echo Running Go tests...
	@for /D %%d in (services\*-service) do ( \
		echo Testing %%d... && \
		cd %%d && go test ./... && cd ..\.. \
	)

build: ## Build all Go services
	@echo Building Go services...
	@for /D %%d in (services\*-service) do ( \
		echo Building %%d... && \
		cd %%d && go build -o bin/server cmd/server/main.go && cd ..\.. \
	)

clean: ## Clean build artifacts
	@echo Cleaning build artifacts...
	@for /D %%d in (services\*-service) do ( \
		if exist %%d\bin rd /s /q %%d\bin \
	)
