package grpc

import (
	"context"

	pb "github.com/homindolenern/goapps-costing-v1/gen/go/costing/v1"
	"github.com/homindolenern/goapps-costing-v1/internal/infrastructure/postgres"
	"github.com/homindolenern/goapps-costing-v1/internal/infrastructure/redis"
)

// HealthHandler implements the gRPC HealthService.
type HealthHandler struct {
	pb.UnimplementedHealthServiceServer
	db    *postgres.DB
	redis *redis.Client
}

// NewHealthHandler creates a new health handler (PostgreSQL only).
func NewHealthHandler(db *postgres.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// NewHealthHandlerWithRedis creates a new health handler with Redis support.
func NewHealthHandlerWithRedis(db *postgres.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// Liveness check - is the service alive and running?
func (h *HealthHandler) Liveness(ctx context.Context, req *pb.LivenessRequest) (*pb.LivenessResponse, error) {
	return &pb.LivenessResponse{
		Status: "OK",
	}, nil
}

// Readiness check - is the service ready to receive traffic?
func (h *HealthHandler) Readiness(ctx context.Context, req *pb.ReadinessRequest) (*pb.ReadinessResponse, error) {
	components := make(map[string]*pb.ComponentHealth)

	// Check PostgreSQL
	if err := h.db.HealthCheck(ctx); err != nil {
		components["postgres"] = &pb.ComponentHealth{
			Status:  "DOWN",
			Message: err.Error(),
		}
	} else {
		components["postgres"] = &pb.ComponentHealth{
			Status:  "UP",
			Message: "",
		}
	}

	// Check Redis (optional)
	if h.redis != nil {
		if err := h.redis.HealthCheck(ctx); err != nil {
			components["redis"] = &pb.ComponentHealth{
				Status:  "DOWN",
				Message: err.Error(),
			}
		} else {
			components["redis"] = &pb.ComponentHealth{
				Status:  "UP",
				Message: "",
			}
		}
	}

	// Check if any component is down
	allUp := true
	for _, comp := range components {
		if comp.Status == "DOWN" {
			allUp = false
			break
		}
	}

	status := "READY"
	if !allUp {
		status = "NOT_READY"
	}

	return &pb.ReadinessResponse{
		Status:     status,
		Components: components,
	}, nil
}
