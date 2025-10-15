package chat

import (
	"time"

	"chat/internal/app/models"
)

// ChatRow — «плоская» проекция строки таблицы chats
type ChatRow struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (row *ChatRow) Values() []any {
	return []any{
		row.ID, row.CreatedAt, row.UpdatedAt,
	}
}

// ToModel конвертирует ChatRow в доменную модель models.Chat
// ParticipantIDs и Messages должны быть загружены отдельными запросами
func ToModel(r *ChatRow) *models.Chat {
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
func FromModel(m *models.Chat) ChatRow {
	if m == nil {
		return ChatRow{}
	}
	return ChatRow{
		ID:        int64(m.ID),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
