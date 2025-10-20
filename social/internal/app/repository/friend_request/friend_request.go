package friend_request

import (
	"time"

	"social/internal/app/models"
)

// Row — «плоская» проекция строки таблицы friend_requests
type Row struct {
	ID         int64      `db:"id"`
	FromUserID int64      `db:"from_user_id"`
	ToUserID   int64      `db:"to_user_id"`
	Status     int        `db:"status"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at"`
}

func (row *Row) Values() []any {
	return []any{
		row.ID, row.FromUserID, row.ToUserID, row.Status, row.CreatedAt, row.UpdatedAt,
	}
}

// ToModel конвертирует Row в доменную модель models.FriendRequest
func ToModel(r *Row) *models.FriendRequest {
	if r == nil {
		return nil
	}
	return &models.FriendRequest{
		ID:         models.FriendRequestID(r.ID),
		FromUserID: models.UserID(r.FromUserID),
		ToUserID:   models.UserID(r.ToUserID),
		Status:     models.FriendRequestStatus(r.Status),
		CreatedAt:  &r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

// FromModel конвертирует доменную модель в Row (для INSERT/UPDATE)
func FromModel(m *models.FriendRequest) Row {
	if m == nil {
		return Row{}
	}

	var createdAt time.Time
	if m.CreatedAt != nil {
		createdAt = *m.CreatedAt
	}

	return Row{
		ID:         int64(m.ID),
		FromUserID: int64(m.FromUserID),
		ToUserID:   int64(m.ToUserID),
		Status:     int(m.Status),
		CreatedAt:  createdAt,
		UpdatedAt:  m.UpdatedAt,
	}
}
