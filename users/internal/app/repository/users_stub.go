package repository

import (
	"context"
	"strings"
	"sync"

	"users/internal/app/models"
)

type UsersRepositoryStub struct {
	mu          sync.RWMutex
	users       map[int64]*models.UserProfile
	usersByNick map[string]*models.UserProfile
	nextID      int64
}

func NewUsersRepositoryStub() *UsersRepositoryStub {
	return &UsersRepositoryStub{
		users:       make(map[int64]*models.UserProfile),
		usersByNick: make(map[string]*models.UserProfile),
		nextID:      1,
	}
}

func (r *UsersRepositoryStub) SaveUser(ctx context.Context, user *models.UserProfile) (*models.UserProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.usersByNick[user.Nickname]; exists {
		return nil, models.ErrAlreadyExists
	}

	savedUser := *user
	savedUser.UserID = r.nextID
	r.nextID++

	r.users[savedUser.UserID] = &savedUser
	r.usersByNick[savedUser.Nickname] = &savedUser

	return &savedUser, nil
}

func (r *UsersRepositoryStub) UpdateUser(ctx context.Context, user *models.UserProfile) (*models.UserProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.users[user.UserID]
	if !ok {
		return nil, models.ErrNotFound
	}

	if existing.Nickname != user.Nickname {
		if _, exists := r.usersByNick[user.Nickname]; exists {
			return nil, models.ErrAlreadyExists
		}
		delete(r.usersByNick, existing.Nickname)
	}

	updatedUser := *user
	r.users[updatedUser.UserID] = &updatedUser
	r.usersByNick[updatedUser.Nickname] = &updatedUser

	return &updatedUser, nil
}

func (r *UsersRepositoryStub) GetUserByID(ctx context.Context, id int64) (*models.UserProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if user, ok := r.users[id]; ok {
		return user, nil
	}

	return nil, nil
}

func (r *UsersRepositoryStub) GetUserByNickname(ctx context.Context, nickname string) (*models.UserProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if user, ok := r.usersByNick[nickname]; ok {
		return user, nil
	}

	return nil, nil
}

func (r *UsersRepositoryStub) SearchByNickname(ctx context.Context, query string, limit int64) ([]*models.UserProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*models.UserProfile
	query = strings.ToLower(query)

	for _, user := range r.users {
		if strings.Contains(strings.ToLower(user.Nickname), query) {
			results = append(results, user)
			if int64(len(results)) >= limit {
				break
			}
		}
	}

	return results, nil
}
