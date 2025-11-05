package usecase

import (
	"context"

	"users/internal/app/models"
	"users/internal/app/usecase/dto"
)

// Порты вторичные
type (
	UsersRepository interface {
		SaveUser(ctx context.Context, user *models.UserProfile) (*models.UserProfile, error)
		UpdateUser(ctx context.Context, user *models.UserProfile) (*models.UserProfile, error)
		GetUserByID(ctx context.Context, id string) (*models.UserProfile, error)
		GetUserByNickname(ctx context.Context, nickname string) (*models.UserProfile, error)
		SearchByNickname(ctx context.Context, query string, limit int64) ([]*models.UserProfile, error)
	}
)

type Usecase interface {
	// CreateProfile создание пользователя
	CreateProfile(ctx context.Context, req dto.CreateProfileRequest) (*models.UserProfile, error)

	// UpdateProfile обновление пользователя
	UpdateProfile(ctx context.Context, req dto.UpdateProfileRequest) (*models.UserProfile, error)

	// GetProfileByID получение пользователя по ID
	GetProfileByID(ctx context.Context, id string) (*models.UserProfile, error)

	// GetProfileByNickname получение пользователя по никнейму
	GetProfileByNickname(ctx context.Context, nickname string) (*models.UserProfile, error)

	// SearchByNickname поиск пользователя по никнейму
	SearchByNickname(ctx context.Context, req dto.SearchByNicknameRequest) ([]*models.UserProfile, error)
}

type UsersService struct {
	usersRepo UsersRepository
}

var _ Usecase = (*UsersService)(nil)

func NewUsecase(usersRepo UsersRepository) *UsersService {
	return &UsersService{
		usersRepo: usersRepo,
	}
}
