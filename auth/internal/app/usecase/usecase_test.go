package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"auth/internal/app/models"
	"auth/internal/app/usecase/dto"
	"auth/internal/app/usecase/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name          string
		req           dto.RegisterRequest
		setupMocks    func(context.Context, *mocks.UsersRepositoryMock, *mocks.UsersServiceMock)
		expectedUser  *models.User
		expectedError error
	}{
		{
			name: "успешная регистрация",
			req: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, repo *mocks.UsersRepositoryMock, service *mocks.UsersServiceMock) {
				repo.GetUserByEmailMock.Expect(ctx, "test@example.com").Return(nil, nil)
				repo.SaveUserMock.Expect(ctx, "test@example.com", "password123").Return(&models.User{
					ID:       1,
					Email:    "test@example.com",
					Password: "password123",
				}, nil)
				service.CreateUserMock.Expect(ctx, int64(1), "test@example.com").Return(nil)
			},
			expectedUser: &models.User{
				ID:       1,
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedError: nil,
		},
		{
			name: "пользователь уже существует",
			req: dto.RegisterRequest{
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, repo *mocks.UsersRepositoryMock, service *mocks.UsersServiceMock) {
				repo.GetUserByEmailMock.Expect(ctx, "existing@example.com").Return(&models.User{
					ID:    1,
					Email: "existing@example.com",
				}, nil)
			},
			expectedUser:  nil,
			expectedError: models.ErrAlreadyExists,
		},
		{
			name: "ошибка при проверке существования пользователя",
			req: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, repo *mocks.UsersRepositoryMock, service *mocks.UsersServiceMock) {
				repo.GetUserByEmailMock.Expect(ctx, "test@example.com").Return(nil, errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("database error"),
		},
		{
			name: "ошибка при сохранении пользователя",
			req: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, repo *mocks.UsersRepositoryMock, service *mocks.UsersServiceMock) {
				repo.GetUserByEmailMock.Expect(ctx, "test@example.com").Return(nil, nil)
				repo.SaveUserMock.Expect(ctx, "test@example.com", "password123").Return(nil, errors.New("save error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("save error"),
		},
		{
			name: "ошибка при создании пользователя в Users Service",
			req: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, repo *mocks.UsersRepositoryMock, service *mocks.UsersServiceMock) {
				repo.GetUserByEmailMock.Expect(ctx, "test@example.com").Return(nil, nil)
				repo.SaveUserMock.Expect(ctx, "test@example.com", "password123").Return(&models.User{
					ID:       1,
					Email:    "test@example.com",
					Password: "password123",
				}, nil)
				service.CreateUserMock.Expect(ctx, int64(1), "test@example.com").Return(errors.New("service error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("service error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			repoMock := mocks.NewUsersRepositoryMock(t)
			serviceMock := mocks.NewUsersServiceMock(t)

			tt.setupMocks(ctx, repoMock, serviceMock)

			authService := NewUsecase(serviceMock, repoMock)

			user, err := authService.Register(ctx, tt.req)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedUser.ID, user.ID)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
				assert.Equal(t, tt.expectedUser.Password, user.Password)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name          string
		req           dto.LoginRequest
		setupMocks    func(context.Context, *mocks.UsersRepositoryMock)
		validateUser  func(*testing.T, *models.User)
		expectedError error
	}{
		{
			name: "успешный логин",
			req: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, repo *mocks.UsersRepositoryMock) {
				repo.GetUserByEmailMock.Expect(ctx, "test@example.com").Return(&models.User{
					ID:       1,
					Email:    "test@example.com",
					Password: "password123",
				}, nil)
				repo.UpdateUserMock.Set(func(ctx context.Context, user *models.User) error {
					require.NotNil(t, user.Token)
					require.NotEmpty(t, user.Token.AccessToken)
					require.NotEmpty(t, user.Token.RefreshToken)
					require.True(t, user.Token.TokenExpiresAt.After(time.Now()))
					return nil
				})
			},
			validateUser: func(t *testing.T, user *models.User) {
				assert.Equal(t, int64(1), user.ID)
				assert.Equal(t, "test@example.com", user.Email)
				require.NotNil(t, user.Token)
				assert.NotEmpty(t, user.Token.AccessToken)
				assert.NotEmpty(t, user.Token.RefreshToken)
				assert.True(t, user.Token.TokenExpiresAt.After(time.Now()))
			},
			expectedError: nil,
		},
		{
			name: "пользователь не найден",
			req: dto.LoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, repo *mocks.UsersRepositoryMock) {
				repo.GetUserByEmailMock.Expect(ctx, "notfound@example.com").Return(nil, nil)
			},
			validateUser:  nil,
			expectedError: models.ErrNotFound,
		},
		{
			name: "неверный пароль",
			req: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMocks: func(ctx context.Context, repo *mocks.UsersRepositoryMock) {
				repo.GetUserByEmailMock.Expect(ctx, "test@example.com").Return(&models.User{
					ID:       1,
					Email:    "test@example.com",
					Password: "password123",
				}, nil)
			},
			validateUser:  nil,
			expectedError: ErrWrongPassword,
		},
		{
			name: "ошибка получения пользователя",
			req: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, repo *mocks.UsersRepositoryMock) {
				repo.GetUserByEmailMock.Expect(ctx, "test@example.com").Return(nil, errors.New("database error"))
			},
			validateUser:  nil,
			expectedError: errors.New("database error"),
		},
		{
			name: "ошибка обновления токена",
			req: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, repo *mocks.UsersRepositoryMock) {
				repo.GetUserByEmailMock.Expect(ctx, "test@example.com").Return(&models.User{
					ID:       1,
					Email:    "test@example.com",
					Password: "password123",
				}, nil)
				repo.UpdateUserMock.Set(func(ctx context.Context, user *models.User) error {
					return errors.New("update error")
				})
			},
			validateUser:  nil,
			expectedError: errors.New("update error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			repoMock := mocks.NewUsersRepositoryMock(t)

			tt.setupMocks(ctx, repoMock)

			authService := NewUsecase(nil, repoMock)

			user, err := authService.Login(ctx, tt.req)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				tt.validateUser(t, user)
			}
		})
	}
}

func TestAuthService_Refresh(t *testing.T) {
	validRefreshToken := "valid-refresh-token"

	tests := []struct {
		name          string
		req           dto.RefreshRequest
		setupMocks    func(context.Context, *mocks.UsersRepositoryMock)
		validateUser  func(*testing.T, *models.User)
		expectedError error
	}{
		{
			name: "успешное обновление токена",
			req: dto.RefreshRequest{
				UserID:       1,
				RefreshToken: validRefreshToken,
			},
			setupMocks: func(ctx context.Context, repo *mocks.UsersRepositoryMock) {
				repo.GetUserByIDMock.Expect(ctx, int64(1)).Return(&models.User{
					ID:       1,
					Email:    "test@example.com",
					Password: "password123",
					Token: &models.UserToken{
						AccessToken:    "old-access-token",
						RefreshToken:   validRefreshToken,
						TokenExpiresAt: time.Now().Add(time.Hour),
					},
				}, nil)
				repo.UpdateUserMock.Set(func(ctx context.Context, user *models.User) error {
					require.NotNil(t, user.Token)
					require.NotEmpty(t, user.Token.AccessToken)
					require.NotEmpty(t, user.Token.RefreshToken)
					require.True(t, user.Token.TokenExpiresAt.After(time.Now()))
					return nil
				})
			},
			validateUser: func(t *testing.T, user *models.User) {
				assert.Equal(t, int64(1), user.ID)
				assert.Equal(t, "test@example.com", user.Email)
				require.NotNil(t, user.Token)
				assert.NotEmpty(t, user.Token.AccessToken)
				assert.NotEmpty(t, user.Token.RefreshToken)
				assert.NotEqual(t, "old-access-token", user.Token.AccessToken)
				assert.True(t, user.Token.TokenExpiresAt.After(time.Now()))
			},
			expectedError: nil,
		},
		{
			name: "пользователь не найден",
			req: dto.RefreshRequest{
				UserID:       999,
				RefreshToken: validRefreshToken,
			},
			setupMocks: func(ctx context.Context, repo *mocks.UsersRepositoryMock) {
				repo.GetUserByIDMock.Expect(ctx, int64(999)).Return(nil, nil)
			},
			validateUser:  nil,
			expectedError: models.ErrNotFound,
		},
		{
			name: "неверный refresh token",
			req: dto.RefreshRequest{
				UserID:       1,
				RefreshToken: "invalid-token",
			},
			setupMocks: func(ctx context.Context, repo *mocks.UsersRepositoryMock) {
				repo.GetUserByIDMock.Expect(ctx, int64(1)).Return(&models.User{
					ID:       1,
					Email:    "test@example.com",
					Password: "password123",
					Token: &models.UserToken{
						AccessToken:    "access-token",
						RefreshToken:   validRefreshToken,
						TokenExpiresAt: time.Now().Add(time.Hour),
					},
				}, nil)
			},
			validateUser:  nil,
			expectedError: ErrWrongToken,
		},
		{
			name: "ошибка получения пользователя",
			req: dto.RefreshRequest{
				UserID:       1,
				RefreshToken: validRefreshToken,
			},
			setupMocks: func(ctx context.Context, repo *mocks.UsersRepositoryMock) {
				repo.GetUserByIDMock.Expect(ctx, int64(1)).Return(nil, errors.New("database error"))
			},
			validateUser:  nil,
			expectedError: errors.New("database error"),
		},
		{
			name: "ошибка обновления токена",
			req: dto.RefreshRequest{
				UserID:       1,
				RefreshToken: validRefreshToken,
			},
			setupMocks: func(ctx context.Context, repo *mocks.UsersRepositoryMock) {
				repo.GetUserByIDMock.Expect(ctx, int64(1)).Return(&models.User{
					ID:       1,
					Email:    "test@example.com",
					Password: "password123",
					Token: &models.UserToken{
						AccessToken:    "access-token",
						RefreshToken:   validRefreshToken,
						TokenExpiresAt: time.Now().Add(time.Hour),
					},
				}, nil)
				repo.UpdateUserMock.Set(func(ctx context.Context, user *models.User) error {
					return errors.New("update error")
				})
			},
			validateUser:  nil,
			expectedError: errors.New("update error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			repoMock := mocks.NewUsersRepositoryMock(t)

			tt.setupMocks(ctx, repoMock)

			authService := NewUsecase(nil, repoMock)

			user, err := authService.Refresh(ctx, tt.req)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				tt.validateUser(t, user)
			}
		})
	}
}
