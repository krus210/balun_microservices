package adapters

import (
	"context"

	pb "auth/pkg/users/api"
)

type UsersClient struct {
	client pb.UsersServiceClient
}

func NewUsersClient(client pb.UsersServiceClient) *UsersClient {
	return &UsersClient{client: client}
}

// CreateProfile - Создание профиля пользователя
func (c *UsersClient) CreateUser(ctx context.Context, userID string, nickname string) error {
	_, err := c.client.CreateProfile(ctx, &pb.CreateProfileRequest{
		UserId:   userID,
		Nickname: nickname,
	})
	if err != nil {
		return err
	}

	return nil
}
