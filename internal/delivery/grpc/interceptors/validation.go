package interceptors

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pb "github.com/homindolenern/goapps-costing-v1/gen/go/costing/v1"
	pkgerrors "github.com/homindolenern/goapps-costing-v1/pkg/errors"
)

// Validation returns a unary server interceptor for protovalidate.
func Validation(validator protovalidate.Validator) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Check if request implements proto.Message
		msg, ok := req.(proto.Message)
		if !ok {
			return handler(ctx, req)
		}

		// Validate the request
		if err := validator.Validate(msg); err != nil {
			// Parse protovalidate error into structured format
			validationErrors := parseProtovalidateError(err)

			// Create structured response
			baseResponse := &pb.BaseResponse{
				StatusCode:       "400",
				IsSuccess:        false,
				Message:          "Validation failed",
				ValidationErrors: validationErrors,
			}

			// Serialize to JSON for error details
			details, marshalErr := json.Marshal(baseResponse)
			if marshalErr != nil {
				return nil, status.Error(codes.InvalidArgument, "Validation failed")
			}

			return nil, status.Error(codes.InvalidArgument, string(details))
		}

		return handler(ctx, req)
	}
}

// parseProtovalidateError parses protovalidate error into structured format.
func parseProtovalidateError(err error) []*pb.ValidationError {
	if err == nil {
		return nil
	}

	// Use errors.As for proper wrapped error handling
	var validationErr *protovalidate.ValidationError
	if errors.As(err, &validationErr) {
		return parseValidationError(validationErr)
	}

	// Fallback: parse error message
	return parseErrorMessage(err.Error())
}

// parseValidationError parses protovalidate.ValidationError.
func parseValidationError(ve *protovalidate.ValidationError) []*pb.ValidationError {
	validationErrors := make([]*pb.ValidationError, 0)

	for _, violation := range ve.Violations {
		field := ""
		message := ""

		// Get field name from FieldDescriptor
		if violation.FieldDescriptor != nil {
			field = string(violation.FieldDescriptor.Name())
		}

		// Get message from Proto if available
		if violation.Proto != nil {
			message = violation.Proto.GetMessage()
		}

		// Fallback: use String() representation for message
		if message == "" {
			message = violation.String()
		}

		validationErrors = append(validationErrors, &pb.ValidationError{
			Field:   field,
			Message: message,
		})
	}

	return validationErrors
}

// parseErrorMessage is a fallback parser for error messages.
func parseErrorMessage(errMsg string) []*pb.ValidationError {
	validationErrors := make([]*pb.ValidationError, 0)

	// Try to parse "validation error: field: message" pattern
	if strings.Contains(errMsg, "validation error:") {
		parts := strings.Split(errMsg, "validation error:")
		for _, part := range parts[1:] {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			// Split "field: message"
			colonIdx := strings.Index(part, ":")
			if colonIdx > 0 {
				field := strings.TrimSpace(part[:colonIdx])
				message := strings.TrimSpace(part[colonIdx+1:])
				validationErrors = append(validationErrors, &pb.ValidationError{
					Field:   field,
					Message: message,
				})
			} else {
				validationErrors = append(validationErrors, &pb.ValidationError{
					Field:   "unknown",
					Message: part,
				})
			}
		}
	}

	if len(validationErrors) == 0 {
		validationErrors = append(validationErrors, &pb.ValidationError{
			Field:   "request",
			Message: errMsg,
		})
	}

	return validationErrors
}

// ParseValidationErrors is a helper to convert pkgerrors to pb.
func ParseValidationErrors(ve *pkgerrors.ValidationErrors) []*pb.ValidationError {
	if ve == nil {
		return nil
	}

	validationErrors := make([]*pb.ValidationError, 0, len(ve.Errors))
	for _, e := range ve.Errors {
		validationErrors = append(validationErrors, &pb.ValidationError{
			Field:   e.Field,
			Message: e.Message,
		})
	}
	return validationErrors
}
