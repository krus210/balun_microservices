package usecase

import (
	"context"
	"fmt"

	"chat/internal/app/models"
	"chat/internal/app/usecase/dto"
)

func (c *ChatService) ListUserChats(ctx context.Context, req dto.ListUserChatsDto) ([]*models.Chat, error) {
	// Проверяем существование пользователя
	userExists, err := c.usersService.CheckUserExists(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("[ChatService][ListUserChats] usersService CheckUserExists error: %w", err)
	}
	if !userExists {
		return nil, models.ErrNotFound
	}

	chats, err := c.chatRepo.ListChatsByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("[ChatService][ListUserChats] chatRepo ListChatsByUserID error: %w", err)
	}

	return chats, nil
}
