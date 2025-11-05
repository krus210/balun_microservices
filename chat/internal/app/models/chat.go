package models

import "time"

type UserID string

type ChatID string

type MessageID string

type Message struct {
	ID        MessageID
	Text      string
	ChatID    ChatID
	OwnerID   UserID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Chat struct {
	ID             ChatID
	ParticipantIDs []UserID
	Messages       []Message
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
