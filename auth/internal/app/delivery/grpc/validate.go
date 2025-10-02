package grpc

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *AuthController) validateCredentials(email string, password string) error {
	var violations []*errdetails.BadRequest_FieldViolation
	if len(email) == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "email",
			Description: "empty",
		})
	}
	if len(password) == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "password",
			Description: "empty",
		})
	}

	if len(violations) > 0 {
		rpcErr := status.New(codes.InvalidArgument, "почта или пароль пустые")

		detailedError, err := rpcErr.WithDetails(&errdetails.BadRequest{
			FieldViolations: violations,
		})
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		return detailedError.Err()
	}

	return nil
}
