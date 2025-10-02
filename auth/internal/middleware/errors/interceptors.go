package errors

import (
	"context"
	"errors"

	"auth/internal/app/models"
	"auth/internal/usecase"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorsUnaryInterceptor - convert any arror to rpc error
func ErrorsUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		//
		if _, ok := status.FromError(err); ok {
			return resp, err
		}

		switch {
		case errors.Is(err, models.ErrNotFound):
			err = status.Error(codes.NotFound, err.Error())
		case errors.Is(err, models.ErrAlreadyExists):
			err = status.Error(codes.AlreadyExists, err.Error())
		case errors.Is(err, usecase.ErrWrongPassword) || errors.Is(err, usecase.ErrWrongToken):
			err = status.Error(codes.Unauthenticated, err.Error())
		default:
			err = status.Error(codes.Unknown, err.Error())
		}

		return resp, err
	}
}
