package grpc

import (
	"context"
	"log"

	"chat/internal/app/models"
	"chat/internal/app/usecase/dto"

	pb "chat/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *ChatController) ListMessages(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	response, err := h.usecase.ListMessages(ctx, dto.ListMessagesDto{
		UserID: "1", // TODO: брать из хедера
		ChatID: models.ChatID(req.ChatId),
		Limit:  req.Limit,
		Cursor: req.Cursor,
	})
	if err != nil {
		return nil, err
	}

	return &pb.ListMessagesResponse{
		Messages:   newPbMessagesFromMessages(response.Messages),
		NextCursor: response.NextCursor,
	}, nil
}
