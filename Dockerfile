# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install required tools
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/master-service \
    ./cmd/master-service

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/master-service /app/master-service

# Copy OpenAPI spec for Swagger UI
COPY --from=builder /app/gen/openapi /app/gen/openapi

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

# Expose ports (gRPC and HTTP)
EXPOSE 9090 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/v1/health/live || exit 1

# Run the binary
ENTRYPOINT ["/app/master-service"]
