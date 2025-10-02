package repository

import (
	"context"
	"sync"

	"auth/internal/app/models"
)

type UsersRepositoryStub struct {
	mu     sync.RWMutex
	users  map[int64]*models.User
	emails map[string]*models.User
	nextID int64
}

func NewUsersRepositoryStub() *UsersRepositoryStub {
	return &UsersRepositoryStub{
		users:  make(map[int64]*models.User),
		emails: make(map[string]*models.User),
		nextID: 1,
	}
}

func (r *UsersRepositoryStub) SaveUser(ctx context.Context, email, password string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.emails[email]; exists {
		return nil, models.ErrAlreadyExists
	}

	user := &models.User{
		ID:       r.nextID,
		Email:    email,
		Password: password,
	}
	r.nextID++

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
		return nil, models.ErrNotFound
	}

	return user, nil
}

func (r *UsersRepositoryStub) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[userID]
	if !exists {
		return nil, models.ErrNotFound
	}

	return user, nil
}
