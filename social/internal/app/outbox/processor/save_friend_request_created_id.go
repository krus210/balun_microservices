package processor

import (
	"context"
	"time"

	"social/internal/app/models"

	"github.com/google/uuid"
)

func (p *Processor) SaveFriendRequestCreatedID(ctx context.Context, id models.FriendRequestID) error {
	e := Event{
		ID:            uuid.New(),
		AggregateType: AggregateTypeFriendRequest,
		AggregateID:   string(id),
		EventType:     EventTypeFriendRequestCreated,
		Payload:       nil,
		CreatedAt:     time.Now().UTC(),
	}
	return p.Repository.SaveEvent(ctx, &e)
}
