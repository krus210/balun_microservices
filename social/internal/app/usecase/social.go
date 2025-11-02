package usecase

import (
	"context"

	"social/internal/app/models"
	"social/internal/app/usecase/dto"
)

// Порты вторичные
type (
	UsersService interface {
		CheckUserExists(ctx context.Context, id models.UserID) (bool, error)
	}

	SocialRepository interface {
		SaveFriendRequest(ctx context.Context, req *models.FriendRequest) (*models.FriendRequest, error)
		UpdateFriendRequest(ctx context.Context, requestID models.FriendRequestID, status models.FriendRequestStatus) (*models.FriendRequest, error)
		GetFriendRequest(ctx context.Context, requestID models.FriendRequestID) (*models.FriendRequest, error)
		GetFriendRequestsByToUserID(ctx context.Context, toUserID models.UserID, limit *int64, cursor *string) (friends []*models.FriendRequest, nextCursor *string, err error)
		GetFriendRequestsByFromUserID(ctx context.Context, fromUserID models.UserID, limit *int64, cursor *string) (friends []*models.FriendRequest, nextCursor *string, err error)
		GetFriendRequestByUserIDs(ctx context.Context, fromUserID models.UserID, toUserID models.UserID) (*models.FriendRequest, error)
		DeleteFriendRequest(ctx context.Context, requestID models.FriendRequestID) error
	}

	// OutboxRepository - репозиторий outbox
	OutboxRepository interface {
		// SaveFriendRequestCreatedID - запись в Outbox сообщения по заказу
		SaveFriendRequestCreatedID(ctx context.Context, id models.FriendRequestID) error
		SaveFriendRequestUpdatedID(ctx context.Context, id models.FriendRequestID, status models.FriendRequestStatus) error
	}

	// TransactionManager
	TransactionManager interface {
		RunReadCommitted(ctx context.Context, f func(ctx context.Context) error) error
	}
)

type Usecase interface {
	// SendFriendRequest отправка заявки на друзья
	SendFriendRequest(ctx context.Context, req dto.FriendRequestDto) (*models.FriendRequest, error)
	// ListFriendRequests получение списка заявок на друзья
	ListFriendRequests(ctx context.Context, toUserID models.UserID) ([]*models.FriendRequest, error)
	// AcceptFriendRequest принятие заявки на друзья
	AcceptFriendRequest(ctx context.Context, req dto.ChangeFriendRequestDto) (*models.FriendRequest, error)
	// DeclineFriendRequest отказ от заявки на друзья
	DeclineFriendRequest(ctx context.Context, req dto.ChangeFriendRequestDto) (*models.FriendRequest, error)
	// RemoveFriend удаление друга
	RemoveFriend(ctx context.Context, req dto.FriendRequestDto) error
	// ListFriends получение списка друзей
	ListFriends(ctx context.Context, req dto.ListFriendsDto) (*dto.ListFriendsResponse, error)
}

type SocialService struct {
	usersService         UsersService
	socialRepo           SocialRepository
	outboxRepository     OutboxRepository
	transactionalManager TransactionManager
}

var _ Usecase = (*SocialService)(nil)

func NewUsecase(
	service UsersService,
	repo SocialRepository,
	outboxRepo OutboxRepository,
	manager TransactionManager,
) *SocialService {
	return &SocialService{
		usersService:         service,
		socialRepo:           repo,
		outboxRepository:     outboxRepo,
		transactionalManager: manager,
	}
}
