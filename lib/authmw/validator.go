package authmw

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken    = errors.New("invalid token")
	ErrTokenExpired    = errors.New("token expired")
	ErrMissingKID      = errors.New("missing kid in token header")
	ErrInvalidAudience = errors.New("invalid audience")
)

// Claims представляет JWT claims
type Claims struct {
	Issuer    string   `json:"iss"`
	Subject   string   `json:"sub"` // user_id
	Audience  []string `json:"aud"`
	IssuedAt  int64    `json:"iat"`
	ExpiresAt int64    `json:"exp"`
	NotBefore int64    `json:"nbf"`
	JWTID     string   `json:"jti"`
}

// JWKSProvider интерфейс для получения JWK по KID
// Реализуется как JWKSCache (HTTP), так и JWKSCacheGRPC
type JWKSProvider interface {
	GetJWKS() *JWKS
	GetKeyByKID(kid string) (*InternalJWK, error)
	Stop()
}

// Validator валидирует JWT токены
type Validator struct {
	cache            JWKSProvider
	expectedIssuer   string
	expectedAudience string // Ожидаемый audience для этого сервиса
}

// ValidatorConfig конфигурация для Validator
type ValidatorConfig struct {
	JWKSCache        JWKSProvider // Может быть *JWKSCache или *JWKSCacheGRPC
	ExpectedIssuer   string       // Например, "balun-auth-service"
	ExpectedAudience string       // Например, "users", "social", "chat"
}

// NewValidator создает новый JWT validator
func NewValidator(cfg ValidatorConfig) *Validator {
	return &Validator{
		cache:            cfg.JWKSCache,
		expectedIssuer:   cfg.ExpectedIssuer,
		expectedAudience: cfg.ExpectedAudience,
	}
}

// Validate валидирует JWT токен и возвращает claims
func (v *Validator) Validate(ctx context.Context, tokenString string) (*Claims, error) {
	// Парсим токен без валидации для получения KID
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Получаем KID из header
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, ErrMissingKID
		}

		// Получаем публичный ключ из кеша
		jwk, err := v.cache.GetKeyByKID(kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get key by KID: %w", err)
		}

		// Конвертируем JWK в RSA public key
		publicKey, err := jwk.ToRSAPublicKey()
		if err != nil {
			return nil, fmt.Errorf("failed to convert JWK to RSA public key: %w", err)
		}

		return publicKey, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}),
		jwt.WithIssuer(v.expectedIssuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Извлекаем claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Валидируем audience
	if v.expectedAudience != "" {
		audInterface, ok := claims["aud"]
		if !ok {
			return nil, ErrInvalidAudience
		}

		// aud может быть string или []string
		var audiences []string
		switch aud := audInterface.(type) {
		case string:
			audiences = []string{aud}
		case []interface{}:
			for _, a := range aud {
				if audStr, ok := a.(string); ok {
					audiences = append(audiences, audStr)
				}
			}
		default:
			return nil, ErrInvalidAudience
		}

		// Проверяем, что наш audience присутствует
		found := false
		for _, aud := range audiences {
			if aud == v.expectedAudience {
				found = true
				break
			}
		}
		if !found {
			return nil, ErrInvalidAudience
		}
	}

	// Конвертируем в нашу структуру Claims
	return mapToClaims(claims)
}

func mapToClaims(claims jwt.MapClaims) (*Claims, error) {
	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid sub claim")
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid jti claim")
	}

	iss, _ := claims["iss"].(string)
	iat, _ := claims["iat"].(float64)
	exp, _ := claims["exp"].(float64)
	nbf, _ := claims["nbf"].(float64)

	result := &Claims{
		Issuer:    iss,
		Subject:   sub,
		IssuedAt:  int64(iat),
		ExpiresAt: int64(exp),
		NotBefore: int64(nbf),
		JWTID:     jti,
	}

	// Парсим audience
	if audInterface, ok := claims["aud"]; ok {
		switch aud := audInterface.(type) {
		case string:
			result.Audience = []string{aud}
		case []interface{}:
			for _, a := range aud {
				if audStr, ok := a.(string); ok {
					result.Audience = append(result.Audience, audStr)
				}
			}
		}
	}

	return result, nil
}
