package usecase

import (
	"context"
	"fmt"

	"chat/internal/app/models"
	"chat/internal/app/usecase/dto"
)

const (
	apiListUserChats = "[ChatService][ListUserChats]"
)

func (c *ChatService) ListUserChats(ctx context.Context, req dto.ListUserChatsDto) ([]*models.Chat, error) {
	// Проверяем существование пользователя
	userExists, err := c.usersService.CheckUserExists(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: usersService CheckUserExists error: %w", apiListUserChats, err)
	}
	if !userExists {
		return nil, models.ErrNotFound
	}

	chats, err := c.chatRepo.ListChatsByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: chatRepo ListChatsByUserID error: %w", apiListUserChats, err)
	}

	return chats, nil
}
