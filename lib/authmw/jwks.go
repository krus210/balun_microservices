package authmw

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
)

// InternalJWK представляет JSON Web Key в формате RSA (внутреннее представление)
// Используется для работы с RSA ключами, отличается от protobuf JWK
type InternalJWK struct {
	KTY string `json:"kty"` // Key Type (RSA)
	Use string `json:"use"` // Public Key Use (sig)
	KID string `json:"kid"` // Key ID
	Alg string `json:"alg"` // Algorithm (RS256)
	N   string `json:"n"`   // Modulus (base64url)
	E   string `json:"e"`   // Exponent (base64url)
}

// JWKS представляет набор JWK ключей
type JWKS struct {
	Keys []InternalJWK `json:"keys"`
}

// ToRSAPublicKey конвертирует InternalJWK в RSA public key
func (jwk *InternalJWK) ToRSAPublicKey() (*rsa.PublicKey, error) {
	// Декодируем modulus (N)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Декодируем exponent (E)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Конвертируем в big.Int
	n := new(big.Int).SetBytes(nBytes)

	// Конвертируем exponent в int
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}

// GetKeyByKID возвращает InternalJWK по KID
func (jwks *JWKS) GetKeyByKID(kid string) (*InternalJWK, error) {
	for i := range jwks.Keys {
		if jwks.Keys[i].KID == kid {
			return &jwks.Keys[i], nil
		}
	}
	return nil, fmt.Errorf("key with KID %s not found", kid)
}
