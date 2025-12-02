//go:generate minimock -i .UsersService,.UsersRepository,.Usecase -s _mock.go -o ./mocks -g
package usecase

import (
	"context"
	"errors"
	"time"

	"auth/internal/app/crypto"
	"auth/internal/app/keystore"
	"auth/internal/app/models"
	"auth/internal/app/token"
	"auth/internal/app/usecase/dto"
)

// Порты вторичные
type (
	UsersService interface {
		CreateUser(ctx context.Context, userID string, nickname string) error
	}

	UsersRepository interface {
		CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error)
		UpdateUser(ctx context.Context, user *models.User) error
		GetUserByEmail(ctx context.Context, email string) (*models.User, error)
		GetUserByID(ctx context.Context, userID string) (*models.User, error)
	}

	RefreshTokensRepository interface {
		CreateToken(ctx context.Context, token *models.RefreshToken) error
		GetTokenByJTI(ctx context.Context, jti string) (*models.RefreshToken, error)
		MarkAsUsed(ctx context.Context, jti, replacedByJTI string) error
		RevokeTokenByJTI(ctx context.Context, jti string) error
	}
)

type Usecase interface {
	// Register создание пользователя
	//
	// ErrAlreadyExists
	Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error)

	// Login аутентификация пользователя
	//
	// ErrNotFound, ErrWrongPassword
	Login(ctx context.Context, req dto.LoginRequest) (*models.User, error)

	// Refresh обновление access token
	//
	// ErrNotFound, ErrTokenUsed, ErrTokenExpired
	Refresh(ctx context.Context, req dto.RefreshRequest) (*models.User, error)

	// Logout отзыв refresh токена
	Logout(ctx context.Context, req dto.LogoutRequest) error

	// GetJWKS получение публичных ключей
	GetJWKS(ctx context.Context) (*dto.JWKSResponse, error)
}

var (
	ErrWrongPassword = errors.New("wrong password")
	ErrWrongToken    = errors.New("wrong token")
	ErrTokenUsed     = errors.New("token already used")
	ErrTokenExpired  = errors.New("token expired")
	ErrInvalidToken  = errors.New("invalid token")
)

type Config struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type AuthService struct {
	usersService      UsersService
	usersRepo         UsersRepository
	refreshTokensRepo RefreshTokensRepository
	passwordHasher    crypto.PasswordHasher
	tokenManager      *token.TokenManager
	keyStore          keystore.KeyStore
	cfg               Config
}

var _ Usecase = (*AuthService)(nil)

func NewUsecase(
	usersService UsersService,
	usersRepo UsersRepository,
	refreshTokensRepo RefreshTokensRepository,
	passwordHasher crypto.PasswordHasher,
	tokenManager *token.TokenManager,
	keyStore keystore.KeyStore,
	cfg Config,
) *AuthService {
	return &AuthService{
		usersService:      usersService,
		usersRepo:         usersRepo,
		refreshTokensRepo: refreshTokensRepo,
		passwordHasher:    passwordHasher,
		tokenManager:      tokenManager,
		keyStore:          keyStore,
		cfg:               cfg,
	}
}
