package usecase

import (
	"context"
	"errors"
	"fmt"

	"auth/internal/app/usecase/dto"

	"auth/internal/app/models"
)

const (
	apiRegister = "[AuthService][Register]"
)

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error) {
	// Проверяем, не существует ли уже пользователь
	existingUser, err := s.usersRepo.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, fmt.Errorf("%s: userRepo GetUserByEmail error: %w", apiRegister, err)
	}
	if existingUser != nil {
		return nil, models.ErrAlreadyExists
	}

	// Хешируем пароль
	passwordHash, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to hash password: %w", apiRegister, err)
	}

	// Создаем пользователя в БД
	user, err := s.usersRepo.CreateUser(ctx, req.Email, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("%s: userRepo CreateUser error: %w", apiRegister, err)
	}

	// Создаем профиль в users сервисе
	err = s.usersService.CreateUser(ctx, user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("%s: usersService CreateUser error: %w", apiRegister, err)
	}

	return user, nil
}
