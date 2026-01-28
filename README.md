# Yarn Costing Microservice

gRPC-based microservice for yarn manufacturing costing system using Clean Architecture + DDD.

## Prerequisites

- Go 1.21+
- Protocol Buffers compiler (protoc)
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose

## Quick Start

```bash
# Install development tools
make setup

# Generate proto files
make proto

# Run database migrations
make migrate-up

# Start the service
make run
```

## Project Structure

```
├── cmd/master-service/     # Application entrypoint
├── internal/
│   ├── domain/             # Domain layer (entities, repos interfaces)
│   ├── application/        # Use cases, commands, queries
│   ├── infrastructure/     # Database, cache implementations
│   └── delivery/grpc/      # gRPC handlers
├── proto/                  # Protobuf definitions
├── gen/                    # Generated code
├── migrations/             # Database migrations
└── deployments/            # Docker, K8s configs
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health/live` | GET | Liveness probe |
| `/health/ready` | GET | Readiness probe |
| `/metrics` | GET | Prometheus metrics |
| `/v1/uoms` | CRUD | Unit of Measure management |
| `/v1/parameters` | CRUD | Parameter management |

## Development

See [docs/DEV_RULES.md](docs/DEV_RULES.md) for development guidelines.

## Known Issues

| Issue | Status | Notes |
|-------|--------|-------|
| OpenAPI generation disabled | ⚠️ Pending | grpc-gateway v2.27.5 has a bug with certain message types. Track: [grpc-gateway#4XXX](https://github.com/grpc-ecosystem/grpc-gateway/issues). Re-enable in `buf.gen.yaml` when fixed. |

**Workaround:** Use `grpcurl` or Postman gRPC for API testing until OpenAPI/Swagger is restored.

## License

Private - Internal use only
