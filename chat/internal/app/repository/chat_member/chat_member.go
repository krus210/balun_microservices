package chat_member

import "chat/internal/app/models"

// Row — «плоская» проекция строки таблицы chat_members
type Row struct {
	ChatID int64 `db:"chat_id"`
	UserID int64 `db:"user_id"`
}

func (row *Row) Values() []any {
	return []any{
		row.ChatID, row.UserID,
	}
}

// ToModel конвертирует Row в пару идентификаторов
func ToModel(r *Row) (models.ChatID, models.UserID) {
	return models.ChatID(r.ChatID), models.UserID(r.UserID)
}

// FromModel создает Row из пары идентификаторов
func FromModel(chatID models.ChatID, userID models.UserID) Row {
	return Row{
		ChatID: int64(chatID),
		UserID: int64(userID),
	}
}
