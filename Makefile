.PHONY: help build run stop clean test docker-build docker-run docker-stop docker-clean docker-logs

# Application configuration
APP_NAME := pacman-game
DOCKER_IMAGE := $(APP_NAME):latest
DOCKER_CONTAINER := $(APP_NAME)
PORT := 8080

# Go configuration
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
BINARY_NAME := server
MAIN_PATH := ./cmd/server

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Local development commands
build: ## Build the application binary
	@echo "Building $(APP_NAME)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_NAME)"

run: ## Run the application locally
	@echo "Starting $(APP_NAME)..."
	$(GOCMD) run $(MAIN_PATH)/main.go

test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

coverage: test ## Run tests and show coverage report
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

tidy: ## Clean up go.mod and go.sum
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	$(GOMOD) verify

lint: ## Run linters
	@echo "Running linters..."
	golangci-lint run ./...

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image: $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built successfully"

docker-run: docker-build ## Build and run Docker container
	@echo "Starting Docker container: $(DOCKER_CONTAINER)..."
	docker run -d \
		--name $(DOCKER_CONTAINER) \
		-p $(PORT):$(PORT) \
		-e SERVER_PORT=$(PORT) \
		$(DOCKER_IMAGE)
	@echo "Container started. Access at http://localhost:$(PORT)"

docker-stop: ## Stop and remove Docker container
	@echo "Stopping Docker container: $(DOCKER_CONTAINER)..."
	@docker stop $(DOCKER_CONTAINER) 2>/dev/null || true
	@docker rm $(DOCKER_CONTAINER) 2>/dev/null || true
	@echo "Container stopped"

docker-logs: ## Show Docker container logs
	docker logs -f $(DOCKER_CONTAINER)

docker-shell: ## Open shell in running container
	docker exec -it $(DOCKER_CONTAINER) /bin/sh

docker-clean: docker-stop ## Remove Docker image and container
	@echo "Cleaning Docker artifacts..."
	@docker rmi $(DOCKER_IMAGE) 2>/dev/null || true
	@echo "Cleanup complete"

# Docker Compose commands
compose-up: ## Start services with docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose up -d
	@echo "Services started. Access at http://localhost:$(PORT)"

compose-down: ## Stop services with docker-compose
	@echo "Stopping services..."
	docker-compose down

compose-logs: ## Show docker-compose logs
	docker-compose logs -f

compose-restart: ## Restart services
	docker-compose restart

compose-rebuild: ## Rebuild and restart services
	@echo "Rebuilding services..."
	docker-compose up -d --build

# Cleanup commands
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "Cleanup complete"

clean-all: clean docker-clean ## Clean everything (build artifacts and Docker)
	@echo "Full cleanup complete"

# Development workflow
dev: ## Start development mode (with hot reload if air is installed)
	@if command -v air > /dev/null; then \
		echo "Starting with hot reload..."; \
		air; \
	else \
		echo "Air not installed. Running normally..."; \
		echo "Install air with: go install github.com/cosmtrek/air@latest"; \
		$(MAKE) run; \
	fi

.DEFAULT_GOAL := help

