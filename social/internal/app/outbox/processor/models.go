package processor

import (
	"time"

	"github.com/google/uuid"
)

type AggregateType string

const (
	AggregateTypeFriendRequest AggregateType = "friend_request"
)

type EventType string

const (
	EventTypeFriendRequestCreated EventType = "FriendRequestCreated"
	EventTypeFriendRequestUpdated EventType = "FriendRequestStatusUpdated"
)

type Event struct {
	ID            uuid.UUID
	AggregateType AggregateType
	AggregateID   string
	EventType     EventType
	Payload       []byte
	CreatedAt     time.Time
	PublishedAt   *time.Time
	RetryCount    int
}
