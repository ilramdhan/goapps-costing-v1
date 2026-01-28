package interceptors

import (
	"context"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Validation returns a unary server interceptor for protovalidate
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
			return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
		}

		return handler(ctx, req)
	}
}
