package dto

import "chat/internal/app/models"

type CreateDirectChatDto struct {
	UserID        models.UserID
	ParticipantID models.UserID
}

type GetChatDto struct {
	UserID models.UserID
	ChatID models.ChatID
}

type ListUserChatsDto struct {
	UserID models.UserID
}

type ListChatMembersDto struct {
	UserID models.UserID
	ChatID models.ChatID
}

type SendMessageDto struct {
	UserID models.UserID
	ChatID models.ChatID
	Text   string
}

type ListMessagesDto struct {
	UserID models.UserID
	ChatID models.ChatID
	Limit  int64
	Cursor *string
}

type ListMessagesResponse struct {
	Messages   []*models.Message
	NextCursor *string
}

type StreamMessagesDto struct {
	UserID      models.UserID
	ChatID      models.ChatID
	SinceUnixMs *int64
}
