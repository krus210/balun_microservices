package grpc

import (
	pb "chat/pkg/api"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *ChatController) StreamMessages(req *pb.StreamMessagesRequest, stream pb.ChatService_StreamMessagesServer) error {
	// TODO: реализовать серверный стрим новых сообщений
	return status.Errorf(codes.Unimplemented, "метод StreamMessages пока не реализован")
}
