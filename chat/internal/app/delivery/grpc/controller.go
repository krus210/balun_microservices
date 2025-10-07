package grpc

import (
	"chat/internal/app/usecase"
	pb "chat/pkg/api"
)

type ChatController struct {
	pb.ChatServiceServer
	usecase usecase.Usecase
}

func NewChatController(usecase usecase.Usecase) *ChatController {
	return &ChatController{
		usecase: usecase,
	}
}
