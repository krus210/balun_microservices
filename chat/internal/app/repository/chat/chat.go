package chat

import (
	"time"

	"chat/internal/app/models"
)

// Row — «плоская» проекция строки таблицы chats
type Row struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (row *Row) Values() []any {
	return []any{
		row.ID, row.CreatedAt, row.UpdatedAt,
	}
}

// ToModel конвертирует ChatRow в доменную модель models.Chat
// ParticipantIDs и Messages должны быть загружены отдельными запросами
func ToModel(r *Row) *models.Chat {
	if r == nil {
		return nil
	}
	return &models.Chat{
		ID:             models.ChatID(r.ID),
		ParticipantIDs: make([]models.UserID, 0),
		Messages:       make([]models.Message, 0),
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

// FromModel конвертирует доменную модель в ChatRow (для INSERT/UPDATE)
func FromModel(m *models.Chat) Row {
	if m == nil {
		return Row{}
	}
	return Row{
		ID:        string(m.ID),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
