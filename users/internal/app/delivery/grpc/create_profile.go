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

func (h *UsersController) CreateProfile(ctx context.Context, req *pb.CreateProfileRequest) (*pb.CreateProfileResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	err := h.validateCredentials(req)
	if err != nil {
		return nil, err
	}

	userProfile, err := h.usecase.CreateProfile(ctx, dto.CreateProfileRequest{
		UserID:    req.UserId,
		Nickname:  req.Nickname,
		Bio:       req.Bio,
		AvatarURL: req.AvatarUrl,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateProfileResponse{
		UserProfile: &pb.UserProfile{
			UserId:    userProfile.UserID,
			Nickname:  userProfile.Nickname,
			Bio:       userProfile.Bio,
			AvatarUrl: userProfile.AvatarURL,
		},
	}, nil
}

func (h *UsersController) validateCredentials(req *pb.CreateProfileRequest) error {
	var violations []*errdetails.BadRequest_FieldViolation
	if len(req.Nickname) == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "nickname",
			Description: "empty",
		})
	}

	if len(violations) > 0 {
		rpcErr := status.New(codes.InvalidArgument, "nickname пустой")

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
