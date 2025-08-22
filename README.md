# Person Service

A REST API microservice for managing person data, built with Go, Gin, and PostgreSQL.

## Features

- **REST API** with two endpoints for storing and retrieving person data
- **Input validation** with structured error responses
- **PostgreSQL database** with GORM ORM and automatic migrations
- **Docker containerization** for easy deployment
- **Integration tests** using PostgreSQL
- **Duplicate prevention** using external UUID constraint

## API Endpoints

### Save Person
Creates a new person record.

```http
POST /save
Content-Type: application/json

{
  "external_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "John Doe",
  "email": "john.doe@example.com",
  "date_of_birth": "1990-01-01T12:00:00Z"
}
```

**Response (201):**
```json
{
  "external_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "John Doe",
  "email": "john.doe@example.com",
  "date_of_birth": "1990-01-01T12:00:00Z"
}
```

**Error responses:** 400 (validation), 409 (duplicate external_id), 500 (server error)

### Get Person
Retrieves a person by internal ID.

```http
GET /{id}
```

**Response (200):** Same format as Save Person  
**Error responses:** 400 (invalid ID), 404 (not found), 500 (server error)

### Health Check
```http
GET /health
```

Returns `{"status": "ok"}` for monitoring.

## Quick Start

**Prerequisites:** Docker and Docker Compose

### 1. Start the service
```bash
docker-compose up --build
```

This starts both the API server (port 8080) and PostgreSQL database.

### 2. Test the API
```bash
# Create a person
curl -X POST http://localhost:8080/save \
  -H "Content-Type: application/json" \
  -d '{
    "external_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "email": "john.doe@example.com",
    "date_of_birth": "1990-01-01T12:00:00Z"
  }'

# Get person by ID (use the ID from the response above)
curl http://localhost:8080/1

# Health check
curl http://localhost:8080/health
```

### 3. Stop the service
```bash
docker-compose down
```

## Development Setup

For local development (when you want to modify and debug the Go code):

### Option 1: Full Docker (Recommended for testing)
```bash
docker-compose up --build
```

### Option 2: Local Go + Docker PostgreSQL (For development)
```bash
# Start only PostgreSQL
docker run -d \
  --name postgres \
  -e POSTGRES_DB=persons \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  postgres:15-alpine

# Run the Go service locally  
go run main.go
```

This allows you to modify the Go code and restart quickly while keeping the database in Docker.

## Running Tests

### Option 1: Automatic (Recommended)
```bash
# This starts PostgreSQL, runs tests, then stops PostgreSQL
make test-with-db
```

### Option 2: Manual
```bash
# Start PostgreSQL first
docker-compose up -d postgres

# Wait a moment for it to start
# Windows: timeout /t 5 /nobreak
# Mac/Linux: sleep 5

# Run tests
go test -v ./tests/...

# Stop PostgreSQL when done
docker-compose down
```

### Option 3: With coverage
```bash
# Start PostgreSQL first
docker-compose up -d postgres

# Run tests with coverage
go test -cover ./tests/...
```

**Note:** Tests use PostgreSQL to ensure they match the production environment exactly.

## Building

```bash
# Build binary
go build -o bin/person-service .

# Build Docker image
docker build -t person-service .

# Format code
go fmt ./...
```

## Project Structure

```
person-service/
├── main.go                 # Application entry point
├── handlers/person.go      # HTTP request handlers
├── models/person.go        # Data models and DTOs
├── database/database.go    # Database connection and migration
├── tests/person_test.go    # Integration tests
├── Dockerfile              # Multi-stage Docker build
├── docker-compose.yml      # Development environment
├── go.mod                  # Go dependencies
└── README.md               # Documentation
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://user:password@localhost:5432/persons?sslmode=disable` | PostgreSQL connection string |
| `PORT` | `8080` | HTTP server port |

## Design Notes

- **Clean Architecture**: Handlers → Models → Database layering
- **Error Handling**: Structured JSON error responses with proper HTTP status codes
- **Validation**: Request validation using Gin's binding with detailed error messages
- **Database**: GORM for type-safe queries and automatic schema migrations
- **Testing**: Integration tests using PostgreSQL (matches production exactly)

## Technology Stack

- **Go 1.23** - Programming language
- **Gin** - HTTP web framework
- **GORM** - Database ORM
- **PostgreSQL** - Database
- **Docker** - Containerization
- **Testify** - Testing framework

---

**Ready to run and test immediately. Total setup time: under 2 minutes.**