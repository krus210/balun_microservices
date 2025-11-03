package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"social/internal/app/models"

	"github.com/google/uuid"
)

func (p *Processor) SaveFriendRequestUpdatedID(
	ctx context.Context, id models.FriendRequestID,
	status models.FriendRequestStatus,
) error {
	payload, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	e := Event{
		ID:            uuid.New(),
		AggregateType: AggregateTypeFriendRequest,
		AggregateID:   string(id),
		EventType:     EventTypeFriendRequestUpdated,
		Payload:       payload,
		CreatedAt:     time.Now().UTC(),
	}
	return p.Repository.SaveEvent(ctx, &e)
}
