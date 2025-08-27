# Todo App Admin Service - Makefile
# Provides convenient shortcuts for common development tasks

# Configuration
SCRIPT_DIR := scripts
DB_RESET_SCRIPT := $(SCRIPT_DIR)/reset-db.sh
TEST_RUNNER_SCRIPT := $(SCRIPT_DIR)/run-tests.sh

# Default target
.DEFAULT_GOAL := help

# Color codes for output
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
BOLD := \033[1m
NC := \033[0m

# Help target - shows available commands
.PHONY: help
help:
	@echo "$(BOLD)$(CYAN)🏗️  Todo App Admin Service - Development Commands$(NC)"
	@echo "$(CYAN)=================================================$(NC)"
	@echo ""
	@echo "$(BOLD)Database Management:$(NC)"
	@echo "  $(GREEN)db-status$(NC)     Show database status and record counts"
	@echo "  $(GREEN)db-clean$(NC)      Clean test data (preserves seed data)"
	@echo "  $(GREEN)db-reset$(NC)      Complete database reset (nuclear option)"
	@echo ""
	@echo "$(BOLD)Testing Commands:$(NC)"
	@echo "  $(GREEN)test$(NC)          Run all test suites with database cleanup"
	@echo "  $(GREEN)test-quick$(NC)    Run all tests without database reset"
	@echo "  $(GREEN)test-foundation$(NC)  Run foundation tests only"
	@echo "  $(GREEN)test-repo$(NC)     Run repository tests only"
	@echo "  $(GREEN)test-services$(NC) Run service layer tests only"
	@echo "  $(GREEN)test-integration$(NC) Run integration test chain"
	@echo "  $(GREEN)test-performance$(NC) Run performance benchmarks"
	@echo ""
	@echo "$(BOLD)Development Commands:$(NC)"
	@echo "  $(GREEN)build$(NC)         Build all packages"
	@echo "  $(GREEN)build-test$(NC)    Build and run build verification"
	@echo "  $(GREEN)lint$(NC)          Run code linting (if available)"
	@echo "  $(GREEN)format$(NC)        Format Go code with gofmt and goimports"
	@echo "  $(GREEN)fmt-check$(NC)     Check code formatting without modifying files"
	@echo "  $(GREEN)tidy$(NC)          Clean up go.mod dependencies"
	@echo ""
	@echo "$(BOLD)Development Workflow:$(NC)"
	@echo "  $(GREEN)dev-setup$(NC)     Set up development environment"
	@echo "  $(GREEN)dev-test$(NC)      Quick development test cycle"
	@echo "  $(GREEN)pre-commit$(NC)    Run pre-commit checks"
	@echo ""
	@echo "$(BOLD)Examples:$(NC)"
	@echo "  make test                    # Run full test suite"
	@echo "  make test-services db-clean  # Run service tests then clean DB"
	@echo "  make dev-test               # Quick development cycle"

# Database Management Targets
.PHONY: db-status db-clean db-reset
db-status:
	@$(DB_RESET_SCRIPT) status

db-clean:
	@$(DB_RESET_SCRIPT) clean

db-reset:
	@$(DB_RESET_SCRIPT) reset

# Testing Targets
.PHONY: test test-quick test-foundation test-repo test-services test-integration test-performance
test:
	@$(TEST_RUNNER_SCRIPT) all

test-quick:
	@$(TEST_RUNNER_SCRIPT) all --no-reset

test-foundation:
	@$(TEST_RUNNER_SCRIPT) foundation

test-repo:
	@$(TEST_RUNNER_SCRIPT) repository

test-services:
	@$(TEST_RUNNER_SCRIPT) services

test-integration:
	@$(TEST_RUNNER_SCRIPT) integration

test-performance:
	@$(TEST_RUNNER_SCRIPT) performance

# Development Targets
.PHONY: build build-test lint format fmt-check tidy
build:
	@echo "$(CYAN)🔨 Building all packages...$(NC)"
	@go build ./...
	@echo "$(GREEN)✅ Build successful$(NC)"

build-test:
	@$(TEST_RUNNER_SCRIPT) build

lint:
	@echo "$(CYAN)🔍 Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "$(GREEN)✅ Linting complete$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint not installed, running go vet instead$(NC)"; \
		go vet ./...; \
	fi

format:
	@echo "$(CYAN)📝 Formatting Go code...$(NC)"
	@gofmt -w .
	@if command -v goimports >/dev/null 2>&1; then \
		echo "$(CYAN)📦 Fixing imports...$(NC)"; \
		goimports -local github.com/todo-app/services/admin-service -w .; \
	else \
		echo "$(YELLOW)⚠️  goimports not installed, run: go install golang.org/x/tools/cmd/goimports@latest$(NC)"; \
		go fmt ./...; \
	fi
	@echo "$(GREEN)✅ Code formatted$(NC)"

fmt-check:
	@echo "$(CYAN)🔍 Checking code formatting...$(NC)"
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "$(RED)❌ The following files are not formatted:$(NC)"; \
		gofmt -l .; \
		exit 1; \
	fi
	@if command -v goimports >/dev/null 2>&1; then \
		if [ -n "$$(goimports -local github.com/todo-app/services/admin-service -l .)" ]; then \
			echo "$(RED)❌ The following files have import issues:$(NC)"; \
			goimports -local github.com/todo-app/services/admin-service -l .; \
			exit 1; \
		fi \
	fi
	@echo "$(GREEN)✅ All files are properly formatted$(NC)"

tidy:
	@echo "$(CYAN)🧹 Cleaning up dependencies...$(NC)"
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies cleaned$(NC)"

# Development Workflow Targets
.PHONY: dev-setup dev-test pre-commit
dev-setup: format tidy build
	@echo "$(GREEN)✅ Development environment ready$(NC)"

dev-test: db-clean test-services
	@echo "$(GREEN)✅ Quick development test cycle complete$(NC)"

pre-commit: format lint build test
	@echo "$(GREEN)✅ Pre-commit checks passed$(NC)"

# Utility Targets
.PHONY: clean clean-all install-deps
clean:
	@echo "$(CYAN)🧹 Cleaning build artifacts...$(NC)"
	@go clean ./...
	@rm -rf dist/ build/
	@echo "$(GREEN)✅ Clean complete$(NC)"

clean-all: clean db-clean
	@echo "$(GREEN)✅ Complete cleanup finished$(NC)"

install-deps:
	@echo "$(CYAN)📦 Installing development dependencies...$(NC)"
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.55.2; \
	fi
	@if ! command -v goimports >/dev/null 2>&1; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	@go mod download
	@echo "$(GREEN)✅ Dependencies installed$(NC)"

# Docker-related targets
.PHONY: docker-start docker-stop docker-logs
docker-start:
	@echo "$(CYAN)🐳 Starting PostgreSQL container...$(NC)"
	@docker start todo-postgres-dev || \
		docker run --name todo-postgres-dev \
		-e POSTGRES_PASSWORD=postgres \
		-p 5432:5432 -d postgres:15-alpine
	@echo "$(GREEN)✅ PostgreSQL container started$(NC)"

docker-stop:
	@echo "$(CYAN)🛑 Stopping PostgreSQL container...$(NC)"
	@docker stop todo-postgres-dev
	@echo "$(GREEN)✅ PostgreSQL container stopped$(NC)"

docker-logs:
	@docker logs -f todo-postgres-dev

# Continuous Integration Simulation
.PHONY: ci ci-quick
ci: clean install-deps pre-commit
	@echo "$(BOLD)$(GREEN)🎉 CI simulation passed!$(NC)"

ci-quick: format lint build test-quick
	@echo "$(BOLD)$(GREEN)🎉 Quick CI check passed!$(NC)"

# Ensure scripts are executable when needed
.PHONY: ensure-scripts-executable
ensure-scripts-executable:
	@chmod +x $(DB_RESET_SCRIPT) $(TEST_RUNNER_SCRIPT) 2>/dev/null || true