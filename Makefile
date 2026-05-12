# Project: MICHA - Monorepo Makefile
# Standardized Industry Workflow (Lint -> Test -> Build)

.PHONY: all test lint build clean backend-test backend-lint backend-build frontend-test frontend-lint frontend-build

# Default target
all: build

# --- Common Commands ---

test: backend-test frontend-test
lint: backend-lint frontend-lint
build: lint test backend-build frontend-build

# --- Backend (Go) ---

backend-test:
	@echo "==> Running Backend Tests with Coverage..."
	@cd backend && go test -v -cover ./...

backend-lint:
	@echo "==> Running Backend Linter (golangci-lint)..."
	@if command -v golangci-lint >/dev/null; then \
		cd backend && golangci-lint run ./... || (echo "Linter failed. If you see 'unsupported version: 2', please update golangci-lint to v1.64.0+ (brew upgrade golangci-lint)"; exit 1); \
	else \
		echo "Skipping golangci-lint (not installed). Use 'brew install golangci-lint' or equivalent."; \
	fi

backend-build:
	@echo "==> Building Backend Binary..."
	@cd backend && go build -o bin/api cmd/api/main.go

# --- Frontend (React/Vite) ---

frontend-test:
	@echo "==> Running Frontend Tests with Coverage..."
	@cd frontend && npm run test:coverage -- --run

frontend-lint:
	@echo "==> Running Frontend Linter (eslint)..."
	@cd frontend && npm run lint

frontend-build:
	@echo "==> Building Frontend Production Bundle..."
	@cd frontend && npm run build

# --- Cleanup ---

clean:
	@echo "==> Cleaning Build Artifacts..."
	@rm -rf backend/bin frontend/dist
