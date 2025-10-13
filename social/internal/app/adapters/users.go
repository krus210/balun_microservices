package adapters

import (
	"context"

	"social/internal/app/models"
	pb "social/pkg/users/api"
)

type UsersClient struct {
	client pb.UsersServiceClient
}

func NewUsersClient(client pb.UsersServiceClient) *UsersClient {
	return &UsersClient{client: client}
}

// CheckUserExists - Проверка существования пользователя
func (c *UsersClient) CheckUserExists(ctx context.Context, id models.UserID) (bool, error) {
	userProfile, err := c.client.GetProfileByID(ctx, &pb.GetProfileByIDRequest{UserId: int64(id)})
	if err != nil {
		return false, err
	}

	return userProfile != nil, nil
}
