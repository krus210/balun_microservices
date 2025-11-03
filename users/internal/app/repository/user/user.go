package user

import (
	"time"

	"users/internal/app/models"
)

// Row — «плоская» проекция строки таблицы user_profiles
type Row struct {
	ID        string     `db:"id"`
	Nickname  string     `db:"nickname"`
	Bio       *string    `db:"bio"`
	AvatarURL *string    `db:"avatar_url"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

func (row *Row) Values() []any {
	return []any{
		row.ID, row.Nickname, row.Bio, row.AvatarURL, row.CreatedAt, row.UpdatedAt,
	}
}

// ToModel конвертирует Row в доменную модель models.UserProfile
func ToModel(r *Row) *models.UserProfile {
	if r == nil {
		return nil
	}
	return &models.UserProfile{
		UserID:    r.ID,
		Nickname:  r.Nickname,
		Bio:       r.Bio,
		AvatarURL: r.AvatarURL,
		CreatedAt: &r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// FromModel конвертирует доменную модель в Row (для INSERT/UPDATE)
func FromModel(m *models.UserProfile) Row {
	if m == nil {
		return Row{}
	}

	var createdAt time.Time
	if m.CreatedAt != nil {
		createdAt = *m.CreatedAt
	}

	return Row{
		ID:        m.UserID,
		Nickname:  m.Nickname,
		Bio:       m.Bio,
		AvatarURL: m.AvatarURL,
		CreatedAt: createdAt,
		UpdatedAt: m.UpdatedAt,
	}
}
