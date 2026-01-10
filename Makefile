.PHONY: all build run clean test fmt help

PROJECT_NAME := seg
BUILD_DIR := bin

all: fmt build test

build: ## Build all binaries
	@echo "Building binaries..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/server ./cmd/server
	go build -o $(BUILD_DIR)/seg ./cmd/seg
	go build -o $(BUILD_DIR)/train_crf ./cmd/train_crf

run: ## Run the segmentation server
	@echo "Starting server..."
	go run cmd/server/main.go

dev: ## Run server with race detector
	go run -race cmd/server/main.go

cli: ## Run the CLI interactive mode
	go run cmd/seg/main.go

train: ## Run standalone CRF training
	go run cmd/train_crf/main.go

test: ## Run unit tests
	@echo "Running tests..."
	go test -v ./...

fmt: ## Format code
	go fmt ./...

clean: ## Remove build artifacts
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -f data/*.bak
	@rm -f data/*.log

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
