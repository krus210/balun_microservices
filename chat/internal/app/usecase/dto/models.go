package dto

import "chat/internal/app/models"

type CreateDirectChatDto struct {
	UserID        int64
	ParticipantID int64
}

type GetChatDto struct {
	UserID int64
	ChatID int64
}

type ListUserChatsDto struct {
	UserID int64
}

type ListChatMembersDto struct {
	UserID int64
	ChatID int64
}

type SendMessageDto struct {
	UserID int64
	ChatID int64
	Text   string
}

type ListMessagesDto struct {
	UserID int64
	ChatID int64
	Limit  int64
	Cursor *string
}

type ListMessagesResponse struct {
	Messages   []*models.Message
	NextCursor *string
}

type StreamMessagesDto struct {
	UserID      int64
	ChatID      int64
	SinceUnixMs *int64
}
