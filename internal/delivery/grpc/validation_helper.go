package grpc

import (
	"context"
	"errors"

	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"

	pb "github.com/homindolenern/goapps-costing-v1/gen/go/costing/v1"
)

// ValidationHelper provides validation utilities for handlers.
type ValidationHelper struct {
	validator protovalidate.Validator
}

// NewValidationHelper creates a new validation helper.
func NewValidationHelper(validator protovalidate.Validator) *ValidationHelper {
	return &ValidationHelper{validator: validator}
}

// Validate validates a proto message and returns BaseResponse with validation errors if any.
func (h *ValidationHelper) Validate(_ context.Context, msg proto.Message) *pb.BaseResponse {
	if h.validator == nil {
		return nil // No validator, skip validation
	}

	err := h.validator.Validate(msg)
	if err == nil {
		return nil // No validation errors
	}

	// Parse validation errors
	validationErrors := h.parseValidationError(err)

	return &pb.BaseResponse{
		StatusCode:       "400",
		IsSuccess:        false,
		Message:          "Validation failed",
		ValidationErrors: validationErrors,
	}
}

// parseValidationError parses protovalidate error into structured format.
func (h *ValidationHelper) parseValidationError(err error) []*pb.ValidationError {
	if err == nil {
		return nil
	}

	// Use errors.As for proper wrapped error handling
	var ve *protovalidate.ValidationError
	if errors.As(err, &ve) {
		return h.parseViolations(ve)
	}

	// Fallback: single error
	return []*pb.ValidationError{
		{Field: "request", Message: err.Error()},
	}
}

// parseViolations parses violations from ValidationError.
func (h *ValidationHelper) parseViolations(ve *protovalidate.ValidationError) []*pb.ValidationError {
	validationErrors := make([]*pb.ValidationError, 0, len(ve.Violations))

	for _, violation := range ve.Violations {
		field := ""
		message := ""

		// Get field name from FieldDescriptor
		if violation.FieldDescriptor != nil {
			field = string(violation.FieldDescriptor.Name())
		}

		// Get message from Proto
		if violation.Proto != nil {
			message = violation.Proto.GetMessage()
		}

		// Fallback
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
