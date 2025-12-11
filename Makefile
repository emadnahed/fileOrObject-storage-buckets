# Makefile for File Storage Service
#
# =============================================================================
# WHAT IS A MAKEFILE?
# =============================================================================
# A Makefile automates common development tasks.
# Instead of typing long commands, you run: make command-name
#
# SYNTAX:
#     target: dependencies
#         command (must be indented with TAB, not spaces!)
#
# USAGE:
#     make help    - Show available commands
#     make build   - Build the application
#     make test    - Run tests
#
# LEARNING RESOURCES:
# - Make tutorial: https://makefiletutorial.com/
# - GNU Make manual: https://www.gnu.org/software/make/manual/
# =============================================================================

# Variables (like constants in programming)
# Use them like: $(VARIABLE_NAME)
.PHONY: help build test clean docker-up docker-down docker-logs install lint format

# Default goal (runs when you type just "make")
.DEFAULT_GOAL := help

# Colors for prettier output (ANSI escape codes)
# These work in most terminals (Linux, macOS, Git Bash on Windows)
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m  # No Color (reset)

# =============================================================================
# HELP COMMAND
# =============================================================================
# This target displays all available commands with descriptions.
# It searches for comments with ## and displays them nicely.
help: ## Show this help message
	@echo "$(BLUE)==================================================$(NC)"
	@echo "$(GREEN)File Storage Service - Available Commands$(NC)"
	@echo "$(BLUE)==================================================$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BLUE)==================================================$(NC)"
	@echo "$(GREEN)Examples:$(NC)"
	@echo "  make docker-up    # Start all services"
	@echo "  make build        # Build all microservices"
	@echo "  make test         # Run all tests"
	@echo "$(BLUE)==================================================$(NC)"

# =============================================================================
# GO MODULE MANAGEMENT
# =============================================================================
install: ## Install/update Go dependencies
	@echo "$(BLUE)Installing Go dependencies...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)✓ Dependencies installed$(NC)"

# =============================================================================
# BUILD COMMANDS
# =============================================================================
build: ## Build all services
	@echo "$(BLUE)Building all services...$(NC)"
	@echo "Building API Gateway..."
	@cd services/api-gateway && go build -o ../../bin/api-gateway .
	@echo "Building Auth Service..."
	@cd services/auth-service && go build -o ../../bin/auth-service .
	@echo "Building File Service..."
	@cd services/file-service && go build -o ../../bin/file-service .
	@echo "Building Metadata Service..."
	@cd services/metadata-service && go build -o ../../bin/metadata-service .
	@echo "Building Processing Service..."
	@cd services/processing-service && go build -o ../../bin/processing-service .
	@echo "Building Versioning Service..."
	@cd services/versioning-service && go build -o ../../bin/versioning-service .
	@echo "Building Notification Service..."
	@cd services/notification-service && go build -o ../../bin/notification-service .
	@echo "$(GREEN)✓ All services built successfully$(NC)"

build-api-gateway: ## Build API Gateway service
	@echo "$(BLUE)Building API Gateway...$(NC)"
	@cd services/api-gateway && go build -o ../../bin/api-gateway .
	@echo "$(GREEN)✓ API Gateway built$(NC)"

build-auth: ## Build Auth service
	@echo "$(BLUE)Building Auth Service...$(NC)"
	@cd services/auth-service && go build -o ../../bin/auth-service .
	@echo "$(GREEN)✓ Auth Service built$(NC)"

# =============================================================================
# TESTING COMMANDS
# =============================================================================
test: ## Run all tests
	@echo "$(BLUE)Running tests...$(NC)"
	go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✓ Tests completed$(NC)"

test-unit: ## Run unit tests only
	@echo "$(BLUE)Running unit tests...$(NC)"
	go test -v -race -short ./...
	@echo "$(GREEN)✓ Unit tests completed$(NC)"

test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(NC)"
	go test -v -tags=integration ./tests/integration/...
	@echo "$(GREEN)✓ Integration tests completed$(NC)"

test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)Generating coverage report...$(NC)"
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"

# =============================================================================
# CODE QUALITY COMMANDS
# =============================================================================
lint: ## Run linter (golangci-lint)
	@echo "$(BLUE)Running linter...$(NC)"
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run --timeout=5m; \
		echo "$(GREEN)✓ Linting completed$(NC)"; \
	else \
		echo "$(YELLOW)⚠ golangci-lint not installed$(NC)"; \
		echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

format: ## Format code with gofmt and goimports
	@echo "$(BLUE)Formatting code...$(NC)"
	gofmt -w -s .
	@if command -v goimports &> /dev/null; then \
		goimports -w .; \
	else \
		echo "$(YELLOW)⚠ goimports not installed, using gofmt only$(NC)"; \
		echo "Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi
	@echo "$(GREEN)✓ Code formatted$(NC)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	go vet ./...
	@echo "$(GREEN)✓ go vet completed$(NC)"

# =============================================================================
# DOCKER COMMANDS
# =============================================================================
docker-up: ## Start all services with Docker Compose
	@echo "$(BLUE)Starting Docker services...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ Services started$(NC)"
	@echo ""
	@echo "$(BLUE)==================================================$(NC)"
	@echo "$(GREEN)Services Available:$(NC)"
	@echo "  MongoDB:      mongodb://localhost:27017"
	@echo "  RabbitMQ:     http://localhost:15672 (guest/guest)"
	@echo "  MinIO:        http://localhost:9001 (minioadmin/minioadmin)"
	@echo "  Redis:        localhost:6379"
	@echo "  Prometheus:   http://localhost:9090"
	@echo "  Grafana:      http://localhost:3000 (admin/admin)"
	@echo "  Jaeger:       http://localhost:16686"
	@echo "$(BLUE)==================================================$(NC)"

docker-down: ## Stop all Docker services
	@echo "$(YELLOW)Stopping Docker services...$(NC)"
	docker-compose down
	@echo "$(GREEN)✓ Services stopped$(NC)"

docker-down-volumes: ## Stop services and remove volumes (WARNING: deletes data!)
	@echo "$(RED)⚠  WARNING: This will delete all data in Docker volumes!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo ""; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose down -v; \
		echo "$(GREEN)✓ Services and volumes removed$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled$(NC)"; \
	fi

docker-restart: ## Restart all Docker services
	@echo "$(BLUE)Restarting Docker services...$(NC)"
	docker-compose restart
	@echo "$(GREEN)✓ Services restarted$(NC)"

docker-logs: ## Show logs from all services
	docker-compose logs -f

docker-logs-mongodb: ## Show MongoDB logs
	docker-compose logs -f mongodb

docker-logs-rabbitmq: ## Show RabbitMQ logs
	docker-compose logs -f rabbitmq

docker-ps: ## Show status of all Docker services
	@echo "$(BLUE)Docker Services Status:$(NC)"
	@docker-compose ps

docker-build: ## Build Docker images
	@echo "$(BLUE)Building Docker images...$(NC)"
	docker-compose build
	@echo "$(GREEN)✓ Docker images built$(NC)"

docker-rebuild: ## Rebuild and restart services
	@echo "$(BLUE)Rebuilding services...$(NC)"
	docker-compose up -d --build
	@echo "$(GREEN)✓ Services rebuilt and restarted$(NC)"

# =============================================================================
# DATABASE COMMANDS
# =============================================================================
db-shell: ## Open MongoDB shell
	@echo "$(BLUE)Opening MongoDB shell...$(NC)"
	docker-compose exec mongodb mongosh -u admin -p changeme file_storage

db-backup: ## Backup MongoDB database
	@echo "$(BLUE)Creating database backup...$(NC)"
	@mkdir -p backups
	docker-compose exec mongodb mongodump --username admin --password changeme --db file_storage --out /tmp/backup
	docker cp go_backend-mongodb-1:/tmp/backup ./backups/mongodb-backup-$$(date +%Y%m%d-%H%M%S)
	@echo "$(GREEN)✓ Database backed up to ./backups/$(NC)"

db-restore: ## Restore MongoDB database from backup (requires BACKUP_DIR)
	@if [ -z "$(BACKUP_DIR)" ]; then \
		echo "$(RED)ERROR: BACKUP_DIR not specified$(NC)"; \
		echo "Usage: make db-restore BACKUP_DIR=./backups/mongodb-backup-20231201"; \
		exit 1; \
	fi
	@echo "$(BLUE)Restoring database from $(BACKUP_DIR)...$(NC)"
	docker cp $(BACKUP_DIR) go_backend-mongodb-1:/tmp/restore
	docker-compose exec mongodb mongorestore --username admin --password changeme --db file_storage /tmp/restore/file_storage
	@echo "$(GREEN)✓ Database restored$(NC)"

# =============================================================================
# DEVELOPMENT COMMANDS
# =============================================================================
dev: docker-up ## Start development environment
	@echo "$(GREEN)✓ Development environment ready$(NC)"
	@echo ""
	@echo "$(BLUE)Next steps:$(NC)"
	@echo "1. Copy .env.example to .env and configure"
	@echo "2. Run: make build"
	@echo "3. Start a service: ./bin/auth-service"

run-auth: ## Run auth service locally
	@echo "$(BLUE)Starting Auth Service...$(NC)"
	go run services/auth-service/main.go

run-gateway: ## Run API gateway locally
	@echo "$(BLUE)Starting API Gateway...$(NC)"
	go run services/api-gateway/main.go

# =============================================================================
# CLEANUP COMMANDS
# =============================================================================
clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache -testcache
	@echo "$(GREEN)✓ Cleanup completed$(NC)"

clean-docker: ## Clean Docker resources (images, containers, volumes)
	@echo "$(RED)⚠  WARNING: This will remove all Docker resources!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo ""; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose down -v; \
		docker system prune -af --volumes; \
		echo "$(GREEN)✓ Docker resources cleaned$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled$(NC)"; \
	fi

# =============================================================================
# UTILITY COMMANDS
# =============================================================================
check-env: ## Check if .env file exists
	@if [ -f .env ]; then \
		echo "$(GREEN)✓ .env file exists$(NC)"; \
	else \
		echo "$(YELLOW)⚠  .env file not found$(NC)"; \
		echo "Run: cp .env.example .env"; \
	fi

deps-check: ## Check if all required tools are installed
	@echo "$(BLUE)Checking dependencies...$(NC)"
	@echo ""
	@command -v go >/dev/null 2>&1 && echo "$(GREEN)✓ Go installed$(NC)" || echo "$(RED)✗ Go not installed$(NC)"
	@command -v docker >/dev/null 2>&1 && echo "$(GREEN)✓ Docker installed$(NC)" || echo "$(RED)✗ Docker not installed$(NC)"
	@command -v docker-compose >/dev/null 2>&1 && echo "$(GREEN)✓ Docker Compose installed$(NC)" || echo "$(RED)✗ Docker Compose not installed$(NC)"
	@command -v golangci-lint >/dev/null 2>&1 && echo "$(GREEN)✓ golangci-lint installed$(NC)" || echo "$(YELLOW)⚠ golangci-lint not installed (optional)$(NC)"
	@echo ""

init: ## Initialize project (first-time setup)
	@echo "$(BLUE)==================================================$(NC)"
	@echo "$(GREEN)Initializing File Storage Project$(NC)"
	@echo "$(BLUE)==================================================$(NC)"
	@echo ""
	@echo "$(BLUE)Step 1: Checking dependencies...$(NC)"
	@make deps-check
	@echo ""
	@echo "$(BLUE)Step 2: Creating .env file...$(NC)"
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "$(GREEN)✓ .env file created$(NC)"; \
		echo "$(YELLOW)⚠  Please edit .env and configure your settings$(NC)"; \
	else \
		echo "$(YELLOW)⚠  .env file already exists$(NC)"; \
	fi
	@echo ""
	@echo "$(BLUE)Step 3: Installing Go dependencies...$(NC)"
	@make install
	@echo ""
	@echo "$(BLUE)Step 4: Starting Docker services...$(NC)"
	@make docker-up
	@echo ""
	@echo "$(GREEN)==================================================$(NC)"
	@echo "$(GREEN)✓ Project initialized successfully!$(NC)"
	@echo "$(GREEN)==================================================$(NC)"
	@echo ""
	@echo "$(BLUE)Next steps:$(NC)"
	@echo "1. Edit .env file with your configuration"
	@echo "2. Run: make build"
	@echo "3. Start services individually or use docker-compose"
	@echo ""

# =============================================================================
# NOTES FOR GO BEGINNERS
# =============================================================================
#
# MAKEFILE SYNTAX:
# - Targets (like "build:") are names for groups of commands
# - @ at start of command suppresses echoing the command itself
# - $ for variables ($$for shell variables, $(VAR) for Make variables)
# - Indentation MUST be tabs, not spaces!
#
# USEFUL MAKE COMMANDS:
# - make -n target    : Show what would be executed (dry run)
# - make -B target    : Force rebuild even if up-to-date
# - make target VAR=value : Pass variable to make
#
# COMMON WORKFLOW:
# 1. make init        : First-time setup
# 2. make dev         : Start development environment
# 3. make build       : Build all services
# 4. make test        : Run tests
# 5. make clean       : Clean up artifacts
#
# DEBUGGING:
# - make docker-logs  : View all service logs
# - make docker-ps    : Check service status
# - make db-shell     : Access MongoDB directly
#
# =============================================================================
