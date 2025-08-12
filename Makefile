# Makefile for Exchange Rate Service

# Variables
APP_NAME := exchange-rate-service
VERSION := 1.0.0
DOCKER_IMAGE := $(APP_NAME):$(VERSION)
DOCKER_IMAGE_LATEST := $(APP_NAME):latest

# Go variables
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

.PHONY: help build run test clean docker-build docker-run docker-compose-up docker-compose-down deps tidy fmt vet lint coverage benchmark

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development commands
build: ## Build the application
	@echo "Building $(APP_NAME)..."
	$(GOBUILD) -o $(GOBIN)/$(APP_NAME) ./cmd/server

run: build ## Run the application locally
	@echo "Running $(APP_NAME)..."
	./$(GOBIN)/$(APP_NAME)

test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-short: ## Run tests without long-running tests
	@echo "Running short tests..."
	$(GOTEST) -short -v ./...

coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -cover -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(GOBIN)/$(APP_NAME)
	rm -f coverage.out coverage.html

# Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOGET) ./...

tidy: ## Clean up and download dependencies
	@echo "Tidying modules..."
	$(GOMOD) tidy

# Code quality
fmt: ## Format code
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...

lint: ## Run golangci-lint (requires golangci-lint to be installed)
	@echo "Running golangci-lint..."
	golangci-lint run

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) -t $(DOCKER_IMAGE_LATEST) .

docker-run: docker-build ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(DOCKER_IMAGE_LATEST)

docker-compose-up: ## Start services with docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose up --build

docker-compose-up-bg: ## Start services with docker-compose in background
	@echo "Starting services with docker-compose in background..."
	docker-compose up --build -d

docker-compose-down: ## Stop docker-compose services
	@echo "Stopping docker-compose services..."
	docker-compose down

docker-compose-logs: ## Show docker-compose logs
	@echo "Showing docker-compose logs..."
	docker-compose logs -f

# Monitoring stack
monitoring-up: ## Start with monitoring (Prometheus & Grafana)
	@echo "Starting services with monitoring..."
	docker-compose --profile monitoring up --build

monitoring-down: ## Stop monitoring services
	@echo "Stopping monitoring services..."
	docker-compose --profile monitoring down

# Development workflow
dev-setup: deps tidy fmt ## Setup development environment
	@echo "Development environment setup complete"

check: fmt vet test ## Run all checks (format, vet, test)
	@echo "All checks passed"

ci: tidy fmt vet test coverage ## Run CI pipeline locally
	@echo "CI pipeline completed"

# Quick testing commands
test-utils: ## Test utils package
	$(GOTEST) -v ./internal/utils/...

test-cache: ## Test cache package
	$(GOTEST) -v ./internal/cache/...

test-services: ## Test services package
	$(GOTEST) -v ./internal/services/...

test-handlers: ## Test handlers package
	$(GOTEST) -v ./internal/handlers/...

# API testing (requires curl)
test-api: ## Test API endpoints (service must be running)
	@echo "Testing API endpoints..."
	@echo "Health check:"
	curl -s http://localhost:8080/health | jq .
	@echo "\nSupported currencies:"
	curl -s http://localhost:8080/api/v1/currencies | jq .
	@echo "\nLatest rate USD to INR:"
	curl -s "http://localhost:8080/api/v1/rates/latest?from=USD&to=INR" | jq .
	@echo "\nConvert 100 USD to INR:"
	curl -s "http://localhost:8080/api/v1/convert?from=USD&to=INR&amount=100" | jq .

# Installation helpers
install-tools: ## Install development tools
	@echo "Installing development tools..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Release commands
release: clean deps tidy fmt vet test docker-build ## Build release
	@echo "Release $(VERSION) built successfully"

# Environment info
env: ## Show environment info
	@echo "Go version: $(shell $(GOCMD) version)"
	@echo "Go path: $(shell $(GOCMD) env GOPATH)"
	@echo "Go root: $(shell $(GOCMD) env GOROOT)"
	@echo "Project path: $(GOBASE)"
	@echo "Binary path: $(GOBIN)"

# File watching (requires 'entr' tool)
watch: ## Watch files and run tests on changes
	find . -name "*.go" | entr -c make test

watch-run: ## Watch files and restart server on changes
	find . -name "*.go" | entr -r make run 