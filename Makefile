# Go related variables
BINARY_NAME := main
MAIN_PATH := ./cmd/api
BUILD_DIR := ./build
TMP_DIR := ./tmp

# Build flags
LDFLAGS := -s -w

.DEFAULT_GOAL := help

# Help target
.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: wire
wire:
	@cd internal && wire

# Development targets
.PHONY: dev
dev: wire ## Start development server with hot reload
	@air

# Build targets
.PHONY: build
build: wire ## Build the application
	@go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Testing targets
.PHONY: test
test: ## Run tests
	@go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

.PHONY: test-race
test-race: ## Run tests with race detection
	@go test -v -race ./...

# Cleaning targets
.PHONY: clean
clean: ## Clean build artifacts
	@rm -rf $(BUILD_DIR) $(TMP_DIR) coverage.out coverage.html
	@go clean -cache

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker containers
	@docker compose build

.PHONY: docker-up
docker-up: ## Start Docker containers
	@docker compose up -d

.PHONY: docker-down
docker-down: ## Stop Docker containers
	@docker compose down

.PHONY: docker-restart
docker-restart: docker-down docker-up ## Restart Docker containers

.PHONY: docker-clean
docker-clean: ## Clean Docker containers and images
	@docker compose down --rmi local

.PHONY: docker
docker: docker-clean docker-build docker-up ## Clean, build and start Docker development environment

# Install development dependencies
.PHONY: install-deps
install-deps: ## Install development dependencies
	@go install github.com/air-verse/air@latest
	@go install github.com/google/wire/cmd/wire@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest