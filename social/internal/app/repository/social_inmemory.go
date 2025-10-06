package repository

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"social/internal/app/models"
)

type InMemorySocialRepository struct {
	mu             sync.RWMutex
	friendRequests map[int64]*models.FriendRequest
	nextID         int64
}

func NewInMemorySocialRepository() *InMemorySocialRepository {
	return &InMemorySocialRepository{
		friendRequests: make(map[int64]*models.FriendRequest),
		nextID:         1,
	}
}

func (r *InMemorySocialRepository) SaveFriendRequest(ctx context.Context, req *models.FriendRequest) (*models.FriendRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	req.ID = r.nextID
	r.nextID++
	req.CreatedAt = &now
	req.UpdatedAt = &now

	r.friendRequests[req.ID] = req
	return req, nil
}

func (r *InMemorySocialRepository) UpdateFriendRequest(ctx context.Context, requestID int64, status models.FriendRequestStatus) (*models.FriendRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	req, exists := r.friendRequests[requestID]
	if !exists {
		return nil, fmt.Errorf("friend request with id %d not found", requestID)
	}

	now := time.Now()
	req.Status = status
	req.UpdatedAt = &now

	return req, nil
}

func (r *InMemorySocialRepository) GetFriendRequest(ctx context.Context, requestID int64) (*models.FriendRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	req, exists := r.friendRequests[requestID]
	if !exists {
		return nil, fmt.Errorf("friend request with id %d not found", requestID)
	}

	return req, nil
}

func (r *InMemorySocialRepository) GetFriendRequestsByToUserID(ctx context.Context, toUserID int64, limit *int64, cursor *string) ([]*models.FriendRequest, *string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var requests []*models.FriendRequest
	for _, req := range r.friendRequests {
		if req.ToUserID == toUserID {
			requests = append(requests, req)
		}
	}

	// Сортировка по ID
	sort.Slice(requests, func(i, j int) bool {
		return requests[i].ID < requests[j].ID
	})

	return r.applyPagination(requests, limit, cursor)
}

func (r *InMemorySocialRepository) GetFriendRequestsByFromUserID(ctx context.Context, fromUserID int64, limit *int64, cursor *string) ([]*models.FriendRequest, *string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var requests []*models.FriendRequest
	for _, req := range r.friendRequests {
		if req.FromUserID == fromUserID {
			requests = append(requests, req)
		}
	}

	// Сортировка по ID
	sort.Slice(requests, func(i, j int) bool {
		return requests[i].ID < requests[j].ID
	})

	return r.applyPagination(requests, limit, cursor)
}

func (r *InMemorySocialRepository) GetFriendRequestByUserIDs(ctx context.Context, fromUserID int64, toUserID int64) (*models.FriendRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, req := range r.friendRequests {
		if req.FromUserID == fromUserID && req.ToUserID == toUserID {
			return req, nil
		}
	}

	return nil, fmt.Errorf("friend request from user %d to user %d not found", fromUserID, toUserID)
}

func (r *InMemorySocialRepository) DeleteFriendRequest(ctx context.Context, requestID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.friendRequests[requestID]; !exists {
		return fmt.Errorf("friend request with id %d not found", requestID)
	}

	delete(r.friendRequests, requestID)
	return nil
}

// applyPagination применяет cursor-based пагинацию к списку заявок
func (r *InMemorySocialRepository) applyPagination(requests []*models.FriendRequest, limit *int64, cursor *string) ([]*models.FriendRequest, *string, error) {
	var startIdx int

	// Если есть курсор, находим позицию для начала
	if cursor != nil && *cursor != "" {
		cursorID, err := strconv.ParseInt(*cursor, 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid cursor: %w", err)
		}

		startIdx = len(requests) // По умолчанию за пределами списка
		for i, req := range requests {
			if req.ID > cursorID {
				startIdx = i
				break
			}
		}
	}

	// Применяем лимит
	defaultLimit := int64(20)
	if limit == nil {
		limit = &defaultLimit
	}

	endIdx := startIdx + int(*limit)
	if endIdx > len(requests) {
		endIdx = len(requests)
	}

	result := requests[startIdx:endIdx]

	// Формируем следующий курсор
	var nextCursor *string
	if len(result) > 0 && endIdx < len(requests) {
		cursorStr := strconv.FormatInt(result[len(result)-1].ID, 10)
		nextCursor = &cursorStr
	}

	return result, nextCursor, nil
}
