package friend_request

import (
	"time"

	"social/internal/app/models"
)

// Row — «плоская» проекция строки таблицы friend_requests
type Row struct {
	ID         string     `db:"id"`
	FromUserID string     `db:"from_user_id"`
	ToUserID   string     `db:"to_user_id"`
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
		ID:         string(m.ID),
		FromUserID: string(m.FromUserID),
		ToUserID:   string(m.ToUserID),
		Status:     int(m.Status),
		CreatedAt:  createdAt,
		UpdatedAt:  m.UpdatedAt,
	}
}
