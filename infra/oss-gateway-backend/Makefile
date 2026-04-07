.PHONY: help build run test test-coverage clean docker-build docker-up docker-down

help:
	@echo "Available commands:"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the server"
	@echo "  make test          - Run tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-up     - Start Docker containers"
	@echo "  make docker-down   - Stop Docker containers"

build:
	go build -o oss-gateway main.go

run:
	go run main.go

test:
	go test -v ./test/...

test-coverage:
	@echo "Running tests with coverage..."
	go test ./test/pkg/... -v -coverprofile=coverage.out -coverpkg=./pkg/...
	@echo.
	@echo "Coverage Summary:"
	@echo "View detailed coverage report with: go tool cover -html coverage.out"
	@echo "Or run: go tool cover -func coverage.out"

clean:
	@if exist oss-gateway.exe del oss-gateway.exe
	@if exist oss-gateway del oss-gateway
	@if exist coverage.out del coverage.out
	go clean

docker-build:
	docker build -t oss-gateway:latest .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

install:
	go mod download
	go mod tidy
