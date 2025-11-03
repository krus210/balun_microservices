package adapters

import (
	"context"

	"chat/internal/app/models"
	pb "chat/pkg/users/api"
)

type UsersClient struct {
	client pb.UsersServiceClient
}

func NewUsersClient(client pb.UsersServiceClient) *UsersClient {
	return &UsersClient{client: client}
}

// CheckUserExists - Проверка существования пользователя
func (c *UsersClient) CheckUserExists(ctx context.Context, id models.UserID) (bool, error) {
	userProfile, err := c.client.GetProfileByID(ctx, &pb.GetProfileByIDRequest{UserId: string(id)})
	if err != nil {
		return false, err
	}

	return userProfile != nil, nil
}
