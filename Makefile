.PHONY: help setup proto build run test lint migrate-up migrate-down clean

# Variables
PROTO_DIR := proto
GEN_DIR := gen
GO_MODULE := github.com/homindolenern/goapps-costing-v1
GRPC_PORT := 9090
HTTP_PORT := 8080
POSTGRES_URL := postgres://postgres:postgres@localhost:5432/costing_db?sslmode=disable

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Install required development tools
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Tools installed successfully!"
	@echo "Make sure $(shell go env GOPATH)/bin is in your PATH"

proto: ## Generate Go code from proto files
	@echo "Generating proto files..."
	buf generate
	@echo "Proto generation complete!"

proto-lint: ## Lint proto files
	buf lint

build: ## Build the application
	go build -o bin/master-service ./cmd/master-service

run: ## Run the application
	go run ./cmd/master-service

test: ## Run tests
	go test -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	golangci-lint run ./...

migrate-up: ## Run database migrations up
	migrate -database "$(POSTGRES_URL)" -path migrations/postgres up

migrate-down: ## Rollback last migration
	migrate -database "$(POSTGRES_URL)" -path migrations/postgres down 1

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_users)
	migrate create -ext sql -dir migrations/postgres -seq $(NAME)

docker-build: ## Build Docker image
	docker build -t costing-master-service:latest -f deployments/docker/Dockerfile .

docker-up: ## Start all services with Docker Compose
	docker-compose -f deployments/docker-compose.yaml up -d

docker-down: ## Stop all services
	docker-compose -f deployments/docker-compose.yaml down

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html
	rm -rf gen/go/* gen/openapi/*

deps: ## Download dependencies
	go mod download
	go mod tidy

.DEFAULT_GOAL := help
