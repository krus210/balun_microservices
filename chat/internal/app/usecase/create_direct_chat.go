package usecase

import (
	"context"
	"fmt"

	"chat/internal/app/models"
	"chat/internal/app/usecase/dto"
)

func (c *ChatService) CreateDirectChat(ctx context.Context, req dto.CreateDirectChatDto) (*models.Chat, error) {
	// Проверяем существование участника
	participantExists, err := c.usersService.CheckUserExists(ctx, req.ParticipantID)
	if err != nil {
		return nil, fmt.Errorf("[ChatService][CreateDirectChat] usersService CheckUserExists error: %w", err)
	}
	if !participantExists {
		return nil, models.ErrNotFound
	}

	// Проверяем, что чат еще не существует
	existingChat, err := c.chatRepo.GetDirectChatByParticipants(ctx, req.UserID, req.ParticipantID)
	if err != nil {
		return nil, fmt.Errorf("[ChatService][CreateDirectChat] chatRepo GetDirectChatByParticipants error: %w", err)
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
		return nil, fmt.Errorf("[ChatService][CreateDirectChat] chatRepo SaveChat error: %w", err)
	}

	return savedChat, nil
}
