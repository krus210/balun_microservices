package repository

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"auth/internal/app/models"
)

type UsersRepositoryStub struct {
	mu     sync.RWMutex
	users  map[string]*models.User
	emails map[string]*models.User
}

func NewUsersRepositoryStub() *UsersRepositoryStub {
	return &UsersRepositoryStub{
		users:  make(map[string]*models.User),
		emails: make(map[string]*models.User),
	}
}

func (r *UsersRepositoryStub) SaveUser(ctx context.Context, email, password string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.emails[email]; exists {
		return nil, models.ErrAlreadyExists
	}

	user := &models.User{
		ID:       uuid.NewString(),
		Email:    email,
		Password: password,
	}

	r.users[user.ID] = user
	r.emails[email] = user

	return user, nil
}

func (r *UsersRepositoryStub) UpdateUser(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return models.ErrNotFound
	}

	r.users[user.ID] = user
	r.emails[user.Email] = user

	return nil
}

func (r *UsersRepositoryStub) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.emails[email]
	if !exists {
		return nil, nil
	}

	return user, nil
}

func (r *UsersRepositoryStub) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[userID]
	if !exists {
		return nil, nil
	}

	return user, nil
}
