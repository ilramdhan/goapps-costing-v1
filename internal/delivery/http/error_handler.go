package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/homindolenern/goapps-costing-v1/gen/go/costing/v1"
)

// CustomErrorHandler handles gRPC errors and returns structured JSON responses
func CustomErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	s, ok := status.FromError(err)
	if !ok {
		runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
		return
	}

	// Try to parse our custom validation error format
	msg := s.Message()
	if s.Code() == codes.InvalidArgument && strings.HasPrefix(msg, "{") {
		// This is our structured JSON error from validation interceptor
		var baseResponse pb.BaseResponse
		if jsonErr := json.Unmarshal([]byte(msg), &baseResponse); jsonErr == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"base": baseResponse,
			})
			return
		}
	}

	// Map gRPC codes to HTTP status and create response
	httpStatus := runtime.HTTPStatusFromCode(s.Code())
	statusCode := httpStatusToString(httpStatus)

	response := map[string]interface{}{
		"base": pb.BaseResponse{
			StatusCode:       statusCode,
			IsSuccess:        false,
			Message:          s.Message(),
			ValidationErrors: []*pb.ValidationError{},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(response)
}

func httpStatusToString(status int) string {
	switch status {
	case http.StatusOK:
		return "200"
	case http.StatusCreated:
		return "201"
	case http.StatusBadRequest:
		return "400"
	case http.StatusUnauthorized:
		return "401"
	case http.StatusForbidden:
		return "403"
	case http.StatusNotFound:
		return "404"
	case http.StatusConflict:
		return "409"
	case http.StatusTooManyRequests:
		return "429"
	case http.StatusInternalServerError:
		return "500"
	case http.StatusServiceUnavailable:
		return "503"
	default:
		return "500"
	}
}

// NewServeMux creates a new gRPC-Gateway ServeMux with custom error handling
func NewServeMux() *runtime.ServeMux {
	return runtime.NewServeMux(
		runtime.WithErrorHandler(CustomErrorHandler),
	)
}
