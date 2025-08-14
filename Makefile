# Go related variables
BINARY_NAME := notifier
MAIN_PATH := ./cmd/notifier
BUILD_DIR := ./build
TMP_DIR := ./tmp

# Build flags
LDFLAGS := -s -w
BUILD_FLAGS := -ldflags="$(LDFLAGS)"

# Default target
.DEFAULT_GOAL := help

# Help target
.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
.PHONY: dev
dev: $(TMP_DIR) ## Start development server with hot reload
	go build -o $(TMP_DIR)/main $(MAIN_PATH) && air

# Build targets
.PHONY: build
build: $(BINARY_NAME) ## Build the application

$(BINARY_NAME):
	go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

.PHONY: build-release
build-release: BUILD_FLAGS += -trimpath
build-release: $(BINARY_NAME) ## Build optimized release binary

# Testing targets
.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: test-race
test-race: ## Run tests with race detection
	go test -v -race ./...

# Cleaning targets
.PHONY: clean
clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR) $(TMP_DIR)
	rm -f coverage.out coverage.html
	go clean -cache

.PHONY: clean-all
clean-all: clean docker-clean ## Clean everything including Docker

# Development Docker targets
.PHONY: docker-build
docker-build: ## Build Docker containers
	docker compose build

.PHONY: docker-up
docker-up: ## Start Docker containers
	docker compose up -d

.PHONY: docker-down
docker-down: ## Stop Docker containers
	docker compose down

.PHONY: docker-restart
docker-restart: docker-down docker-up ## Restart Docker containers

.PHONY: docker-clean
docker-clean: ## Clean Docker containers and images
	docker compose down --rmi local

.PHONY: docker-dev
docker-dev: docker-clean docker-build docker-up ## Clean, build and start Docker development environment

# Utility targets
$(TMP_DIR):
	mkdir -p $(TMP_DIR)

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)
