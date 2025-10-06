package usecase

import (
	"context"

	"social/internal/app/models"
	"social/internal/app/usecase/dto"
)

// Порты вторичные
type (
	UsersService interface {
		CheckUserExists(ctx context.Context, id int64) (bool, error)
	}

	SocialRepository interface {
		SaveFriendRequest(ctx context.Context, req *models.FriendRequest) (*models.FriendRequest, error)
		UpdateFriendRequest(ctx context.Context, requestID int64, status models.FriendRequestStatus) (*models.FriendRequest, error)
		GetFriendRequest(ctx context.Context, requestID int64) (*models.FriendRequest, error)
		GetFriendRequestsByToUserID(ctx context.Context, toUserID int64, limit *int64, cursor *string) (friends []*models.FriendRequest, nextCursor *string, err error)
		GetFriendRequestsByFromUserID(ctx context.Context, fromUserID int64, limit *int64, cursor *string) (friends []*models.FriendRequest, nextCursor *string, err error)
		GetFriendRequestByUserIDs(ctx context.Context, fromUserID int64, toUserID int64) (*models.FriendRequest, error)
		DeleteFriendRequest(ctx context.Context, requestID int64) error
	}
)

type Usecase interface {
	// SendFriendRequest отправка заявки на друзья
	SendFriendRequest(ctx context.Context, req dto.FriendRequestDto) (*models.FriendRequest, error)
	// ListFriendRequests получение списка заявок на друзья
	ListFriendRequests(ctx context.Context, toUserID int64) ([]*models.FriendRequest, error)
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
	usersService UsersService
	socialRepo   SocialRepository
}

var _ Usecase = (*SocialService)(nil)

func NewUsecase(service UsersService, repo SocialRepository) *SocialService {
	return &SocialService{
		usersService: service,
		socialRepo:   repo,
	}
}
