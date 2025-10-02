package grpc

import (
	"auth/internal/app/usecase"
	pb "auth/pkg/api"
)

type AuthController struct {
	pb.AuthServiceServer
	usecase usecase.Usecase
}

func NewAuthController(usecase usecase.Usecase) *AuthController {
	return &AuthController{
		usecase: usecase,
	}
}
