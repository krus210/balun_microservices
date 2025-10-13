package usecase

import (
	"context"
	"fmt"

	"chat/internal/app/models"
	"chat/internal/app/usecase/dto"
)

const (
	api = "[ChatService][CreateDirectChat]"
)

func (c *ChatService) CreateDirectChat(ctx context.Context, req dto.CreateDirectChatDto) (*models.Chat, error) {
	// Проверяем существование участника
	participantExists, err := c.usersService.CheckUserExists(ctx, req.ParticipantID)
	if err != nil {
		return nil, fmt.Errorf("%s: usersService CheckUserExists error: %w", api, err)
	}
	if !participantExists {
		return nil, fmt.Errorf("%s: %w", api, models.ErrNotFound)
	}

	// Проверяем, что чат еще не существует
	existingChat, err := c.chatRepo.GetDirectChatByParticipants(ctx, req.UserID, req.ParticipantID)
	if err != nil {
		return nil, fmt.Errorf("%s: chatRepo GetDirectChatByParticipants error: %w", api, err)
	}
	if existingChat != nil {
		return nil, models.ErrAlreadyExists
	}

	// Создаем чат
	chat := &models.Chat{
		ParticipantIDs: []models.UserID{req.UserID, req.ParticipantID},
	}

	savedChat, err := c.chatRepo.SaveChat(ctx, chat)
	if err != nil {
		return nil, fmt.Errorf("%s: chatRepo SaveChat error: %w", api, err)
	}

	return savedChat, nil
}
