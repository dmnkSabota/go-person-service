# Person Service

REST API for managing person data with Go and PostgreSQL.

## Endpoints

- `POST /save` - Create person
- `GET /{id}` - Get person by ID
- `GET /health` - Health check

## Running

```bash
docker-compose up --build
```

## Testing

```bash
curl -X POST http://localhost:8080/save \
  -H "Content-Type: application/json" \
  -d '{"external_id":"550e8400-e29b-41d4-a716-446655440000","name":"John Doe","email":"john@example.com","date_of_birth":"1990-01-01T12:00:00Z"}'

curl http://localhost:8080/1
```

## Development

```bash
docker-compose up -d postgres
go run main.go
go test ./tests/
```

## Structure

- `handlers/` - HTTP handlers
- `models/` - Data models
- `database/` - DB connection
- `tests/` - Integration tests

Uses Gin, GORM, PostgreSQL. External ID prevents duplicates.