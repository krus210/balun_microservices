package message

import (
	"time"

	"chat/internal/app/models"
)

// Row — «плоская» проекция строки таблицы messages
type Row struct {
	ID        int64     `db:"id"`
	Text      string    `db:"text"`
	ChatID    int64     `db:"chat_id"`
	OwnerID   int64     `db:"owner_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (row *Row) Values() []any {
	return []any{
		row.ID, row.Text, row.ChatID, row.OwnerID, row.CreatedAt, row.UpdatedAt,
	}
}

// ToModel конвертирует Row в доменную модель models.Message
func ToModel(r *Row) *models.Message {
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

// FromModel конвертирует доменную модель в Row (для INSERT/UPDATE)
func FromModel(m *models.Message) Row {
	if m == nil {
		return Row{}
	}
	return Row{
		ID:        int64(m.ID),
		Text:      m.Text,
		ChatID:    int64(m.ChatID),
		OwnerID:   int64(m.OwnerID),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
