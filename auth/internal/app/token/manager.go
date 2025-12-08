package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"auth/internal/app/keystore"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

// RefreshTokenClaims - claims для refresh токена
type RefreshTokenClaims struct {
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"` // user_id
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	JWTID     string `json:"jti"`
	Type      string `json:"type"` // "refresh"
	DeviceID  string `json:"device_id,omitempty"`
}

// TokenManager - менеджер для создания и валидации JWT токенов
type TokenManager struct {
	cfg      Config
	keyStore keystore.KeyStore
}

func NewTokenManager(cfg Config, keyStore keystore.KeyStore) *TokenManager {
	return &TokenManager{
		cfg:      cfg,
		keyStore: keyStore,
	}
}

// CreateAccessToken - создает access JWT токен
func (tm *TokenManager) CreateAccessToken(ctx context.Context, userID string) (string, error) {
	// Получаем активный ключ для подписи
	key, err := tm.keyStore.GetActiveKey(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get active key: %w", err)
	}

	privateKey, err := keystore.DecodePrivateKeyFromPEM(key.PrivateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss": tm.cfg.Issuer,
		"sub": userID,
		"aud": tm.cfg.Audience,
		"iat": now.Unix(),
		"exp": now.Add(tm.cfg.AccessTokenTTL).Unix(),
		"nbf": now.Unix(),
		"jti": uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = key.KID

	return token.SignedString(privateKey)
}

// CreateRefreshToken - создает refresh JWT токен
func (tm *TokenManager) CreateRefreshToken(ctx context.Context, userID, deviceID string) (string, string, error) {
	key, err := tm.keyStore.GetActiveKey(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get active key: %w", err)
	}

	privateKey, err := keystore.DecodePrivateKeyFromPEM(key.PrivateKeyPEM)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode private key: %w", err)
	}

	now := time.Now()
	jti := uuid.New().String()

	claims := jwt.MapClaims{
		"iss":  tm.cfg.Issuer,
		"sub":  userID,
		"iat":  now.Unix(),
		"exp":  now.Add(tm.cfg.RefreshTokenTTL).Unix(),
		"jti":  jti,
		"type": "refresh",
	}

	if deviceID != "" {
		claims["device_id"] = deviceID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = key.KID

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, jti, nil
}

// VerifyRefreshToken - верифицирует refresh токен
func (tm *TokenManager) VerifyRefreshToken(ctx context.Context, tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.Parse(tokenString, tm.keyFunc(ctx),
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}),
		jwt.WithIssuer(tm.cfg.Issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, ErrInvalidToken
	}

	return tm.mapToRefreshClaims(claims)
}

func (tm *TokenManager) keyFunc(ctx context.Context) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("JWT missing 'kid' header")
		}

		key, err := tm.keyStore.GetKeyByKID(ctx, kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get key by KID: %w", err)
		}

		return keystore.DecodePublicKeyFromPEM(key.PublicKeyPEM)
	}
}

func (tm *TokenManager) mapToRefreshClaims(claims jwt.MapClaims) (*RefreshTokenClaims, error) {
	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, ErrInvalidToken
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, ErrInvalidToken
	}

	result := &RefreshTokenClaims{
		Issuer:    tm.cfg.Issuer,
		Subject:   sub,
		IssuedAt:  int64(iat),
		ExpiresAt: int64(exp),
		JWTID:     jti,
		Type:      "refresh",
	}

	if deviceID, ok := claims["device_id"].(string); ok {
		result.DeviceID = deviceID
	}

	return result, nil
}
