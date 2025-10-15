package message

import (
	"time"

	"chat/internal/app/models"
)

// MessageRow — «плоская» проекция строки таблицы messages
type MessageRow struct {
	ID        int64     `db:"id"`
	Text      string    `db:"text"`
	ChatID    int64     `db:"chat_id"`
	OwnerID   int64     `db:"owner_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (row *MessageRow) Values() []any {
	return []any{
		row.ID, row.Text, row.ChatID, row.OwnerID, row.CreatedAt, row.UpdatedAt,
	}
}

// ToMessageModel конвертирует MessageRow в доменную модель models.Message
func ToMessageModel(r *MessageRow) *models.Message {
	if r == nil {
		return nil
	}
	return &models.Message{
		ID:        models.MessageID(r.ID),
		Text:      r.Text,
		ChatID:    models.ChatID(r.ChatID),
		OwnerID:   models.UserID(r.OwnerID),
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// FromMessageModel конвертирует доменную модель в MessageRow (для INSERT/UPDATE)
func FromMessageModel(m *models.Message) MessageRow {
	if m == nil {
		return MessageRow{}
	}
	return MessageRow{
		ID:        int64(m.ID),
		Text:      m.Text,
		ChatID:    int64(m.ChatID),
		OwnerID:   int64(m.OwnerID),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
