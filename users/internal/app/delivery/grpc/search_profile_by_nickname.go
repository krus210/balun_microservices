package grpc

import (
	"context"
	"log"

	"users/internal/app/usecase/dto"

	pb "users/pkg/api"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (h *UsersController) SearchProfileByNickname(ctx context.Context, req *pb.SearchByNicknameRequest) (*pb.SearchByNicknameResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	err := h.validateQuery(req)
	if err != nil {
		return nil, err
	}

	userProfiles, err := h.usecase.SearchByNickname(ctx, dto.SearchByNicknameRequest{
		Query: req.Query,
		Limit: req.Limit,
	})
	if err != nil {
		return nil, err
	}

	return &pb.SearchByNicknameResponse{
		Results: newPbUserProfilesFromUserProfiles(userProfiles),
	}, nil
}

func (h *UsersController) validateQuery(req *pb.SearchByNicknameRequest) error {
	var violations []*errdetails.BadRequest_FieldViolation
	if len(req.Query) == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "query",
			Description: "empty",
		})
	}
	if req.Limit == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "limit",
			Description: "zero",
		})
	}

	if len(violations) > 0 {
		rpcErr := status.New(codes.InvalidArgument, "query пустой или limit = 0")

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
