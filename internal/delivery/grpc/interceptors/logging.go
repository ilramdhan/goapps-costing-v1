package interceptors

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Logging returns a unary server interceptor for logging.
func Logging() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Get status code
		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			} else {
				code = codes.Unknown
			}
		}

		// Log the request
		duration := time.Since(start)
		logger := log.With().
			Str("method", info.FullMethod).
			Dur("duration", duration).
			Str("code", code.String()).
			Logger()

		if err != nil {
			logger.Error().Err(err).Msg("gRPC request failed")
		} else {
			logger.Info().Msg("gRPC request completed")
		}

		return resp, err
	}
}
