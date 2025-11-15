package interceptors

import (
	"context"

	"github.com/sskorolev/balun_microservices/lib/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LogErrorUnaryInterceptor - log interceptor
func LogErrorUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		logCtx := logger.ToContext(ctx,
			logger.FromContext(ctx).With(
				"operation", info.FullMethod,
				"component", "middleware",
			),
		)

		logger.Debug(logCtx, "receive request")
		resp, err = handler(logCtx, req)
		logger.Debug(logCtx, "handle request")

		if err != nil {
			// 4ХХ -> warn
			// 5ХХ -> Error
			logger.ErrorKV(logCtx, "server error",
				"rpc_error", err.Error(),
				"rpc_code", status.Code(err),
			)
		}

		return resp, err
	}
}
