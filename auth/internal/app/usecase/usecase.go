//go:generate minimock -i .UsersService,.UsersRepository,.Usecase -s _mock.go -o ./mocks -g
package usecase

import (
	"context"
	"errors"

	"auth/internal/app/usecase/dto"

	"auth/internal/app/models"
)

// Порты вторичные
type (
	UsersService interface {
		CreateUser(ctx context.Context, userID int64, nickname string) error
	}

	UsersRepository interface {
		SaveUser(ctx context.Context, email, password string) (*models.User, error)
		UpdateUser(ctx context.Context, user *models.User) error
		GetUserByEmail(ctx context.Context, email string) (*models.User, error)
		GetUserByID(ctx context.Context, userID int64) (*models.User, error)
	}
)

type Usecase interface {
	// Register создание пользователя
	//
	// ErrAlreadyExists
	Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error)

	// Login аутентификация пользователя
	//
	// ErrNotFound
	Login(ctx context.Context, req dto.LoginRequest) (*models.User, error)

	// Refresh обновление access token
	//
	// ErrNotFound
	Refresh(ctx context.Context, req dto.RefreshRequest) (*models.User, error)
}

var (
	ErrWrongPassword = errors.New("wrong password")
	ErrWrongToken    = errors.New("wrong token")
)

type AuthService struct {
	usersService UsersService
	usersRepo    UsersRepository
}

var _ Usecase = (*AuthService)(nil)

func NewUsecase(usersService UsersService, usersRepo UsersRepository) *AuthService {
	return &AuthService{
		usersService: usersService,
		usersRepo:    usersRepo,
	}
}
