package processor

import (
	"context"
	"strconv"
	"time"

	"social/internal/app/models"

	"github.com/google/uuid"
)

func (p *Processor) SaveFriendRequestCreatedID(ctx context.Context, id models.FriendRequestID) error {
	e := Event{
		ID:            uuid.New(),
		AggregateType: AggregateTypeFriendRequest,
		AggregateID:   strconv.FormatInt(int64(id), 10),
		EventType:     EventTypeFriendRequestCreated,
		Payload:       nil,
		CreatedAt:     time.Now().UTC(),
	}
	return p.Repository.SaveEvent(ctx, &e)
}
