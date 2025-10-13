package grpc

import (
	"users/internal/app/usecase"
	pb "users/pkg/api"
)

type UsersController struct {
	pb.UsersServiceServer
	usecase usecase.Usecase
}

func NewUsersController(usecase usecase.Usecase) *UsersController {
	return &UsersController{
		usecase: usecase,
	}
}
