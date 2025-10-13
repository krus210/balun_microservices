package models

import "time"

type UserID int64

type ChatID int64

type MessageID int64

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
