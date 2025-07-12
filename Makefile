# Makefile for jonbu-ohlcv

.PHONY: build test clean run-server run-cli docker-up docker-down migrate-up migrate-down

# Build the application
build:
	@echo "Building jonbu-ohlcv..."
	@./scripts/build.sh

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean

# Run the server
run-server:
	@echo "Starting server..."
	@go run cmd/server/main.go

# Run CLI
run-cli:
	@echo "Running CLI..."
	@go run cmd/cli/main.go

# Docker operations
docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d

docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down

docker-build:
	@echo "Building Docker images..."
	@docker-compose build

# Database migrations (when implemented)
migrate-up:
	@echo "Running database migrations..."
	@go run cmd/cli/main.go migrate up

migrate-down:
	@echo "Rolling back database migrations..."
	@go run cmd/cli/main.go migrate down

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run

# Development setup
setup:
	@echo "Setting up development environment..."
	@cp config/.env.example config/.env
	@echo "Please edit config/.env with your configuration"

# Show help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  run-server    - Run the HTTP server"
	@echo "  run-cli       - Run the CLI"
	@echo "  docker-up     - Start Docker services"
	@echo "  docker-down   - Stop Docker services"
	@echo "  docker-build  - Build Docker images"
	@echo "  migrate-up    - Run database migrations"
	@echo "  migrate-down  - Rollback database migrations"
	@echo "  deps          - Install dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  setup         - Setup development environment"
	@echo "  help          - Show this help message"
