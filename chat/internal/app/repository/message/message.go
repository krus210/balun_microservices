package message

import (
	"time"

	"chat/internal/app/models"
)

// Row — «плоская» проекция строки таблицы messages
type Row struct {
	ID        string    `db:"id"`
	Text      string    `db:"text"`
	ChatID    string    `db:"chat_id"`
	OwnerID   string    `db:"owner_id"`
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
		ID:        string(m.ID),
		Text:      m.Text,
		ChatID:    string(m.ChatID),
		OwnerID:   string(m.OwnerID),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
