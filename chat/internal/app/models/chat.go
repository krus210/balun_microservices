package models

import "time"

type Message struct {
	ID        int64
	Text      string
	ChatID    int64
	OwnerID   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Chat struct {
	ID             int64
	ParticipantIDs []int64
	Messages       []Message
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
