package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"auth/internal/app/models"
	"auth/internal/app/usecase/dto"
	"auth/internal/app/usecase/mocks"
	pb "auth/pkg/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthController_Register(t *testing.T) {
	tests := []struct {
		name             string
		req              *pb.RegisterRequest
		setupMocks       func(context.Context, *mocks.UsecaseMock)
		expectedResponse *pb.RegisterResponse
		expectedError    error
		checkError       func(*testing.T, error)
	}{
		{
			name: "успешная регистрация",
			req: &pb.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, usecase *mocks.UsecaseMock) {
				usecase.RegisterMock.Expect(ctx, dto.RegisterRequest{
					Email:    "test@example.com",
					Password: "password123",
				}).Return(&models.User{
					ID:       "550e8400-e29b-41d4-a716-446655440000",
					Email:    "test@example.com",
					Password: "password123",
				}, nil)
			},
			expectedResponse: &pb.RegisterResponse{
				UserId: "550e8400-e29b-41d4-a716-446655440000",
			},
			expectedError: nil,
		},
		{
			name: "пустой email",
			req: &pb.RegisterRequest{
				Email:    "",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, usecase *mocks.UsecaseMock) {
				// мок не вызывается, ошибка валидации
			},
			expectedResponse: nil,
			checkError: func(t *testing.T, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "пустой password",
			req: &pb.RegisterRequest{
				Email:    "test@example.com",
				Password: "",
			},
			setupMocks: func(ctx context.Context, usecase *mocks.UsecaseMock) {
				// мок не вызывается, ошибка валидации
			},
			expectedResponse: nil,
			checkError: func(t *testing.T, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "ошибка от usecase - пользователь уже существует",
			req: &pb.RegisterRequest{
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, usecase *mocks.UsecaseMock) {
				usecase.RegisterMock.Expect(ctx, dto.RegisterRequest{
					Email:    "existing@example.com",
					Password: "password123",
				}).Return(nil, models.ErrAlreadyExists)
			},
			expectedResponse: nil,
			expectedError:    models.ErrAlreadyExists,
		},
		{
			name: "ошибка от usecase - внутренняя ошибка",
			req: &pb.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, usecase *mocks.UsecaseMock) {
				usecase.RegisterMock.Expect(ctx, dto.RegisterRequest{
					Email:    "test@example.com",
					Password: "password123",
				}).Return(nil, errors.New("database error"))
			},
			expectedResponse: nil,
			expectedError:    errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			usecaseMock := mocks.NewUsecaseMock(t)
			tt.setupMocks(ctx, usecaseMock)

			controller := NewAuthController(usecaseMock)

			resp, err := controller.Register(ctx, tt.req)

			if tt.checkError != nil {
				tt.checkError(t, err)
				assert.Nil(t, resp)
			} else if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResponse.UserId, resp.UserId)
			}
		})
	}
}

func TestAuthController_Login(t *testing.T) {
	tests := []struct {
		name             string
		req              *pb.LoginRequest
		setupMocks       func(context.Context, *mocks.UsecaseMock)
		expectedResponse *pb.LoginResponse
		expectedError    error
		checkError       func(*testing.T, error)
	}{
		{
			name: "успешный логин",
			req: &pb.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, usecase *mocks.UsecaseMock) {
				usecase.LoginMock.Expect(ctx, dto.LoginRequest{
					Email:    "test@example.com",
					Password: "password123",
				}).Return(&models.User{
					ID:    "550e8400-e29b-41d4-a716-446655440000",
					Email: "test@example.com",
					Token: &models.UserToken{
						AccessToken:    "access-token",
						RefreshToken:   "refresh-token",
						TokenExpiresAt: time.Now().Add(time.Hour),
					},
				}, nil)
			},
			expectedResponse: &pb.LoginResponse{
				UserId:       "550e8400-e29b-41d4-a716-446655440000",
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
			},
			expectedError: nil,
		},
		{
			name: "пустой email",
			req: &pb.LoginRequest{
				Email:    "",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, usecase *mocks.UsecaseMock) {
				// мок не вызывается, ошибка валидации
			},
			expectedResponse: nil,
			checkError: func(t *testing.T, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "пустой password",
			req: &pb.LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			setupMocks: func(ctx context.Context, usecase *mocks.UsecaseMock) {
				// мок не вызывается, ошибка валидации
			},
			expectedResponse: nil,
			checkError: func(t *testing.T, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "пользователь не найден",
			req: &pb.LoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, usecase *mocks.UsecaseMock) {
				usecase.LoginMock.Expect(ctx, dto.LoginRequest{
					Email:    "notfound@example.com",
					Password: "password123",
				}).Return(nil, models.ErrNotFound)
			},
			expectedResponse: nil,
			expectedError:    models.ErrNotFound,
		},
		{
			name: "неверный пароль",
			req: &pb.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMocks: func(ctx context.Context, usecase *mocks.UsecaseMock) {
				usecase.LoginMock.Expect(ctx, dto.LoginRequest{
					Email:    "test@example.com",
					Password: "wrongpassword",
				}).Return(nil, errors.New("wrong password"))
			},
			expectedResponse: nil,
			expectedError:    errors.New("wrong password"),
		},
		{
			name: "ошибка от usecase",
			req: &pb.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(ctx context.Context, usecase *mocks.UsecaseMock) {
				usecase.LoginMock.Expect(ctx, dto.LoginRequest{
					Email:    "test@example.com",
					Password: "password123",
				}).Return(nil, errors.New("database error"))
			},
			expectedResponse: nil,
			expectedError:    errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			usecaseMock := mocks.NewUsecaseMock(t)
			tt.setupMocks(ctx, usecaseMock)

			controller := NewAuthController(usecaseMock)

			resp, err := controller.Login(ctx, tt.req)

			if tt.checkError != nil {
				tt.checkError(t, err)
				assert.Nil(t, resp)
			} else if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResponse.UserId, resp.UserId)
				assert.Equal(t, tt.expectedResponse.AccessToken, resp.AccessToken)
				assert.Equal(t, tt.expectedResponse.RefreshToken, resp.RefreshToken)
			}
		})
	}
}

func TestAuthController_Refresh(t *testing.T) {
	tests := []struct {
		name             string
		req              *pb.RefreshRequest
		setupMocks       func(context.Context, *mocks.UsecaseMock)
		expectedResponse *pb.RefreshResponse
		expectedError    error
	}{
		{
			name: "успешное обновление токена",
			req: &pb.RefreshRequest{
				RefreshToken: "valid-refresh-token",
			},
			setupMocks: func(ctx context.Context, usecase *mocks.UsecaseMock) {
				usecase.RefreshMock.Expect(ctx, dto.RefreshRequest{
					UserID:       "1", // hardcoded в методе Refresh (пока string "1", потом исправим на реальный UUID)
					RefreshToken: "valid-refresh-token",
				}).Return(&models.User{
					ID:    "550e8400-e29b-41d4-a716-446655440000",
					Email: "test@example.com",
					Token: &models.UserToken{
						AccessToken:    "new-access-token",
						RefreshToken:   "new-refresh-token",
						TokenExpiresAt: time.Now().Add(time.Hour),
					},
				}, nil)
			},
			expectedResponse: &pb.RefreshResponse{
				UserId:       "550e8400-e29b-41d4-a716-446655440000",
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
			},
			expectedError: nil,
		},
		{
			name: "пользователь не найден",
			req: &pb.RefreshRequest{
				RefreshToken: "valid-refresh-token",
			},
			setupMocks: func(ctx context.Context, usecase *mocks.UsecaseMock) {
				usecase.RefreshMock.Expect(ctx, dto.RefreshRequest{
					UserID:       "1",
					RefreshToken: "valid-refresh-token",
				}).Return(nil, models.ErrNotFound)
			},
			expectedResponse: nil,
			expectedError:    models.ErrNotFound,
		},
		{
			name: "неверный refresh token",
			req: &pb.RefreshRequest{
				RefreshToken: "invalid-token",
			},
			setupMocks: func(ctx context.Context, usecase *mocks.UsecaseMock) {
				usecase.RefreshMock.Expect(ctx, dto.RefreshRequest{
					UserID:       "1",
					RefreshToken: "invalid-token",
				}).Return(nil, errors.New("wrong token"))
			},
			expectedResponse: nil,
			expectedError:    errors.New("wrong token"),
		},
		{
			name: "ошибка от usecase",
			req: &pb.RefreshRequest{
				RefreshToken: "valid-refresh-token",
			},
			setupMocks: func(ctx context.Context, usecase *mocks.UsecaseMock) {
				usecase.RefreshMock.Expect(ctx, dto.RefreshRequest{
					UserID:       "1",
					RefreshToken: "valid-refresh-token",
				}).Return(nil, errors.New("database error"))
			},
			expectedResponse: nil,
			expectedError:    errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			usecaseMock := mocks.NewUsecaseMock(t)
			tt.setupMocks(ctx, usecaseMock)

			controller := NewAuthController(usecaseMock)

			resp, err := controller.Refresh(ctx, tt.req)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResponse.UserId, resp.UserId)
				assert.Equal(t, tt.expectedResponse.AccessToken, resp.AccessToken)
				assert.Equal(t, tt.expectedResponse.RefreshToken, resp.RefreshToken)
			}
		})
	}
}
