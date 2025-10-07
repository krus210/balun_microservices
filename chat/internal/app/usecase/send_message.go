package usecase

import (
	"context"
	"fmt"

	"chat/internal/app/models"
	"chat/internal/app/usecase/dto"
)

func (c *ChatService) SendMessage(ctx context.Context, req dto.SendMessageDto) (*models.Message, error) {
	chat, err := c.chatRepo.GetChat(ctx, req.ChatID)
	if err != nil {
		return nil, fmt.Errorf("[ChatService][SendMessage] chatRepo GetChat error: %w", err)
	}
	if chat == nil {
		return nil, models.ErrNotFound
	}

	// Проверяем, что пользователь является участником чата
	isMember, err := c.chatRepo.IsChatMember(ctx, req.ChatID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("[ChatService][SendMessage] chatRepo IsChatMember error: %w", err)
	}
	if !isMember {
		return nil, models.ErrPermissionDenied
	}

	// Создаем сообщение
	message := &models.Message{
		ChatID:  req.ChatID,
		OwnerID: req.UserID,
		Text:    req.Text,
	}

	savedMessage, err := c.chatRepo.SaveMessage(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("[ChatService][SendMessage] chatRepo SaveMessage error: %w", err)
	}

	return savedMessage, nil
}
