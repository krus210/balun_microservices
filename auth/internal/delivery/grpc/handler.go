package grpc

import (
	"auth/internal/usecase"
	pb "auth/pkg/api"
)

type AuthHandler struct {
	pb.AuthServiceServer
	usecases usecase.Usecases
}

func NewAuthHandler(usecases usecase.Usecases) *AuthHandler {
	return &AuthHandler{
		usecases: usecases,
	}
}
