package usecase

import (
	"context"
	"fmt"

	"chat/internal/app/models"
	"chat/internal/app/usecase/dto"
)

func (c *ChatService) ListMessages(ctx context.Context, req dto.ListMessagesDto) (*dto.ListMessagesResponse, error) {
	chat, err := c.chatRepo.GetChat(ctx, req.ChatID)
	if err != nil {
		return nil, fmt.Errorf("[ChatService][GetChat] chatRepo GetChat error: %w", err)
	}
	if chat == nil {
		return nil, models.ErrNotFound
	}

	// Проверяем, что пользователь является участником чата
	isMember, err := c.chatRepo.IsChatMember(ctx, req.ChatID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("[ChatService][ListMessages] chatRepo IsChatMember error: %w", err)
	}
	if !isMember {
		return nil, models.ErrPermissionDenied
	}

	messages, nextCursor, err := c.chatRepo.ListMessages(ctx, req.ChatID, req.Limit, req.Cursor)
	if err != nil {
		return nil, fmt.Errorf("[ChatService][ListMessages] chatRepo ListMessages error: %w", err)
	}

	return &dto.ListMessagesResponse{
		Messages:   messages,
		NextCursor: nextCursor,
	}, nil
}
