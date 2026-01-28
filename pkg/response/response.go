package response

import (
	pb "github.com/homindolenern/goapps-costing-v1/gen/go/costing/v1"
	pkgerrors "github.com/homindolenern/goapps-costing-v1/pkg/errors"
)

// Builder helps construct consistent responses
type Builder struct{}

// New creates a new response builder
func New() *Builder {
	return &Builder{}
}

// Success creates a success base response
func (b *Builder) Success(message string) *pb.BaseResponse {
	return &pb.BaseResponse{
		StatusCode:       "200",
		IsSuccess:        true,
		Message:          message,
		ValidationErrors: []*pb.ValidationError{},
	}
}

// Created creates a created response (201)
func (b *Builder) Created(message string) *pb.BaseResponse {
	return &pb.BaseResponse{
		StatusCode:       "201",
		IsSuccess:        true,
		Message:          message,
		ValidationErrors: []*pb.ValidationError{},
	}
}

// ValidationError creates a validation error response
func (b *Builder) ValidationError(errors *pkgerrors.ValidationErrors) *pb.BaseResponse {
	validationErrors := make([]*pb.ValidationError, 0, len(errors.Errors))
	for _, e := range errors.Errors {
		validationErrors = append(validationErrors, &pb.ValidationError{
			Field:   e.Field,
			Message: e.Message,
		})
	}

	return &pb.BaseResponse{
		StatusCode:       "400",
		IsSuccess:        false,
		Message:          "Validation failed",
		ValidationErrors: validationErrors,
	}
}

// NotFound creates a not found response
func (b *Builder) NotFound(message string) *pb.BaseResponse {
	return &pb.BaseResponse{
		StatusCode:       "404",
		IsSuccess:        false,
		Message:          message,
		ValidationErrors: []*pb.ValidationError{},
	}
}

// Conflict creates a conflict response (already exists)
func (b *Builder) Conflict(message string) *pb.BaseResponse {
	return &pb.BaseResponse{
		StatusCode:       "409",
		IsSuccess:        false,
		Message:          message,
		ValidationErrors: []*pb.ValidationError{},
	}
}

// BadRequest creates a bad request response
func (b *Builder) BadRequest(message string) *pb.BaseResponse {
	return &pb.BaseResponse{
		StatusCode:       "400",
		IsSuccess:        false,
		Message:          message,
		ValidationErrors: []*pb.ValidationError{},
	}
}

// InternalError creates an internal server error response
func (b *Builder) InternalError(message string) *pb.BaseResponse {
	return &pb.BaseResponse{
		StatusCode:       "500",
		IsSuccess:        false,
		Message:          message,
		ValidationErrors: []*pb.ValidationError{},
	}
}

// Unauthorized creates an unauthorized response
func (b *Builder) Unauthorized(message string) *pb.BaseResponse {
	return &pb.BaseResponse{
		StatusCode:       "401",
		IsSuccess:        false,
		Message:          message,
		ValidationErrors: []*pb.ValidationError{},
	}
}

// Forbidden creates a forbidden response
func (b *Builder) Forbidden(message string) *pb.BaseResponse {
	return &pb.BaseResponse{
		StatusCode:       "403",
		IsSuccess:        false,
		Message:          message,
		ValidationErrors: []*pb.ValidationError{},
	}
}

// RateLimited creates a rate limited response
func (b *Builder) RateLimited(message string) *pb.BaseResponse {
	return &pb.BaseResponse{
		StatusCode:       "429",
		IsSuccess:        false,
		Message:          message,
		ValidationErrors: []*pb.ValidationError{},
	}
}

// Global builder instance for convenience
var Default = New()

// Convenience functions using default builder
func Success(message string) *pb.BaseResponse       { return Default.Success(message) }
func Created(message string) *pb.BaseResponse       { return Default.Created(message) }
func NotFound(message string) *pb.BaseResponse      { return Default.NotFound(message) }
func Conflict(message string) *pb.BaseResponse      { return Default.Conflict(message) }
func BadRequest(message string) *pb.BaseResponse    { return Default.BadRequest(message) }
func InternalError(message string) *pb.BaseResponse { return Default.InternalError(message) }
func Unauthorized(message string) *pb.BaseResponse  { return Default.Unauthorized(message) }
func Forbidden(message string) *pb.BaseResponse     { return Default.Forbidden(message) }
func RateLimited(message string) *pb.BaseResponse   { return Default.RateLimited(message) }

func ValidationError(errors *pkgerrors.ValidationErrors) *pb.BaseResponse {
	return Default.ValidationError(errors)
}
