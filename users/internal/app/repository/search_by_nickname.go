package repository

import (
	"context"
	"fmt"

	"users/internal/app/models"
	"users/internal/app/repository/user"

	"github.com/Masterminds/squirrel"
)

const searchByNicknameApi = "[Repository][SearchByNickname]"

// SearchByNickname ищет пользователей по части никнейма
func (r *Repository) SearchByNickname(ctx context.Context, query string, limit int64) ([]*models.UserProfile, error) {
	// Получаем QueryEngine из контекста (может быть транзакция или обычное соединение)
	conn := r.tm.GetQueryEngine(ctx)

	// Запрос для поиска пользователей по никнейму с использованием LIKE
	// Используем % для поиска подстроки в начале никнейма
	searchQuery := r.sb.Select(user.UserProfilesTableColumns...).
		From(user.UserProfilesTable).
		Where(squirrel.Like{user.UserProfilesTableColumnNickname: query + "%"}).
		OrderBy(user.UserProfilesTableColumnNickname).
		Limit(uint64(limit))

	// Выполняем запрос
	var rows []user.Row
	if err := conn.Selectx(ctx, &rows, searchQuery); err != nil {
		return nil, fmt.Errorf("%s: %w", searchByNicknameApi, ConvertPGError(err))
	}

	// Конвертируем строки в модели
	profiles := make([]*models.UserProfile, 0, len(rows))
	for i := range rows {
		profiles = append(profiles, user.ToModel(&rows[i]))
	}

	return profiles, nil
}
