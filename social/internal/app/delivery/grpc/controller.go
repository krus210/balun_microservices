package grpc

import (
	"social/internal/app/usecase"
	pb "social/pkg/api"
)

type SocialController struct {
	pb.SocialServiceServer
	usecase usecase.Usecase
}

func NewSocialController(usecase usecase.Usecase) *SocialController {
	return &SocialController{
		usecase: usecase,
	}
}
