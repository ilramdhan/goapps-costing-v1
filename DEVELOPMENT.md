# Development Rules

## Code Standards

### Go Conventions
1. **Package Names**: Use lowercase, single-word names (e.g., `uom`, `parameter`)
2. **Interfaces**: Suffix with `-er` when possible (e.g., `Reader`, `Writer`)
3. **Error Handling**: Return errors, don't panic in library code
4. **Comments**: All exported functions and types must have doc comments

### Project Structure
```
internal/
├── domain/          # Business logic, entities, value objects
├── application/     # Use cases, commands, queries, handlers
├── delivery/        # gRPC handlers, HTTP handlers
└── infrastructure/  # Database, cache, external services
pkg/                 # Reusable packages
```

### Validation
- Proto validation rules defined in `.proto` files using `buf.build/validate`
- Handler-level validation using `ValidationHelper`
- Domain-level validation in entity constructors

### Response Format
All API responses use `BaseResponse`:
```json
{
  "base": {
    "statusCode": "200",
    "isSuccess": true,
    "message": "Success",
    "validationErrors": []
  },
  "data": {}
}
```

---

## Git Workflow

### Branch Naming
- `feature/` - New features
- `fix/` - Bug fixes
- `refactor/` - Code refactoring
- `docs/` - Documentation

### Commit Messages
Format: `type: description`

Types:
- `feat:` - New feature
- `fix:` - Bug fix
- `refactor:` - Code refactoring
- `docs:` - Documentation
- `test:` - Tests
- `chore:` - Build/tooling

---

## Testing

### Run Tests
```bash
# All tests
go test -v ./...

# Integration tests only
go test -v ./tests/integration/...

# With coverage
go test -v -coverprofile=coverage.out ./...
```

### Test Categories
- **Unit Tests**: `*_test.go` in same package
- **Integration Tests**: `tests/integration/`
- **E2E Tests**: Manual via grpcurl

---

## Development Commands

```bash
# Start dependencies
docker compose -f deployments/docker-compose.yaml up -d

# Run server
make run
# OR
go run ./cmd/master-service

# Run lint
golangci-lint run

# Generate proto
buf generate

# Run migrations
migrate -path migrations/postgres -database "$DATABASE_URL" up
```

---

## API Documentation

- **Swagger UI**: http://localhost:8080/swagger/
- **gRPC Reflection**: Enabled for debugging
- **Metrics**: http://localhost:8080/metrics

---

## Adding New Services

1. **Proto**: Add `proto/{service}/v1/{service}.proto`
2. **Domain**: Create `internal/domain/{service}/`
3. **Application**: Create `internal/application/{service}/`
4. **Handler**: Create `internal/delivery/grpc/{service}_handler.go`
5. **Repository**: Create `internal/infrastructure/postgres/{service}_repository.go`
6. **Migration**: Add `migrations/postgres/{timestamp}_{name}.up.sql`
7. **Test**: Add tests in `tests/integration/`
8. **Register**: Wire everything in `cmd/master-service/main.go`
