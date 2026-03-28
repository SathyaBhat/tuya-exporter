.PHONY: help build test clean docker-build docker-up docker-down fmt lint run

help:
	@echo "Tuya Exporter - Available Commands"
	@echo "==================================="
	@echo "make build        - Build the binary"
	@echo "make test         - Run tests"
	@echo "make clean        - Remove built artifacts"
	@echo "make fmt          - Format code"
	@echo "make lint         - Run linter"
	@echo "make run          - Run locally (requires env vars)"
	@echo "make docker-build - Build Docker image"
	@echo "make docker-up    - Start with docker-compose"
	@echo "make docker-down  - Stop docker-compose services"
	@echo "make docker-logs  - View docker-compose logs"

build:
	@echo "Building tuya-exporter..."
	go build -o tuya-exporter ./cmd/exporter

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning..."
	rm -f tuya-exporter
	go clean

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Running linter..."
	go vet ./...

run: build
	@echo "Running exporter..."
	./tuya-exporter

docker-build:
	@echo "Building Docker image..."
	docker build -t tuya-exporter:latest .

docker-up:
	@echo "Starting services with docker-compose..."
	docker-compose up -d
	@echo "Services started. Run 'make docker-logs' to see logs."

docker-down:
	@echo "Stopping docker-compose services..."
	docker-compose down

docker-logs:
	@echo "Following docker-compose logs..."
	docker-compose logs -f

dev-setup:
	@echo "Setting up development environment..."
	go mod download
	go mod tidy
	@echo "Done. Copy .env.example to .env and configure."

.DEFAULT_GOAL := help
