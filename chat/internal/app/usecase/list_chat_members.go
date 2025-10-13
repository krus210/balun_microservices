package usecase

import (
	"context"
	"fmt"

	"chat/internal/app/models"
	"chat/internal/app/usecase/dto"
)

func (c *ChatService) ListChatMembers(ctx context.Context, req dto.ListChatMembersDto) ([]models.UserID, error) {
	chat, err := c.chatRepo.GetChat(ctx, req.ChatID)
	if err != nil {
		return nil, fmt.Errorf("[ChatService][ListChatMembers] chatRepo GetChat error: %w", err)
	}
	if chat == nil {
		return nil, models.ErrNotFound
	}

	// Проверяем, что пользователь является участником чата
	isMember, err := c.chatRepo.IsChatMember(ctx, req.ChatID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("[ChatService][ListChatMembers] chatRepo IsChatMember error: %w", err)
	}
	if !isMember {
		return nil, models.ErrPermissionDenied
	}

	members, err := c.chatRepo.GetChatMembers(ctx, req.ChatID)
	if err != nil {
		return nil, fmt.Errorf("[ChatService][ListChatMembers] chatRepo GetChatMembers error: %w", err)
	}

	return members, nil
}
