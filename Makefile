.PHONY: build run test test-with-db test-coverage docker-up docker-down clean fmt deps help

# Build the application
build:
	go build -o bin/person-service .

# Run the application
run:
	go run main.go

# Run tests (requires PostgreSQL running)
test:
	go test -v ./tests/...

# Start PostgreSQL and run tests
test-with-db:
	@echo "Starting PostgreSQL..."
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@timeout /t 8 /nobreak >nul 2>&1 || sleep 8
	@echo "Running tests..."
	go test -v ./tests/...
	@echo "Stopping PostgreSQL..."
	docker-compose down

# Run tests with coverage
test-coverage:
	go test -cover ./tests/...

# Start with Docker Compose
docker-up:
	docker-compose up --build

# Stop Docker Compose
docker-down:
	docker-compose down

# Clean build artifacts
clean:
	rm -rf bin/

# Format code
fmt:
	go fmt ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Show help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application binary"
	@echo "  run           - Run the application locally"
	@echo "  test          - Run tests (requires PostgreSQL running)"
	@echo "  test-with-db  - Start PostgreSQL, run tests, stop PostgreSQL"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  docker-up     - Start with Docker Compose (app + database)"
	@echo "  docker-down   - Stop Docker Compose"
	@echo "  clean         - Clean build artifacts"
	@echo "  fmt           - Format Go code"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  help          - Show this help message"