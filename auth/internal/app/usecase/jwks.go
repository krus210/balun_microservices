package usecase

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"fmt"

	"auth/internal/app/keystore"
	"auth/internal/app/usecase/dto"
)

const (
	apiGetJWKS = "[AuthService][GetJWKS]"
)

func (s *AuthService) GetJWKS(ctx context.Context) (*dto.JWKSResponse, error) {
	// Получаем все активные ключи
	keys, err := s.keyStore.GetAllActiveKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get active keys: %w", apiGetJWKS, err)
	}

	jwks := make([]dto.JWK, 0, len(keys))

	for _, key := range keys {
		// Декодируем публичный ключ из PEM
		publicKey, err := keystore.DecodePublicKeyFromPEM(key.PublicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to decode public key: %w", apiGetJWKS, err)
		}

		// Конвертируем RSA public key в JWK формат
		jwk, err := rsaPublicKeyToJWK(key.KID, publicKey)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to convert key to JWK: %w", apiGetJWKS, err)
		}

		jwks = append(jwks, jwk)
	}

	return &dto.JWKSResponse{Keys: jwks}, nil
}

// rsaPublicKeyToJWK конвертирует RSA public key в JWK формат
func rsaPublicKeyToJWK(kid string, publicKey *rsa.PublicKey) (dto.JWK, error) {
	// Modulus (n) - base64url encoding
	nBytes := publicKey.N.Bytes()
	n := base64.RawURLEncoding.EncodeToString(nBytes)

	// Exponent (e) - base64url encoding
	eBytes := make([]byte, 4)
	eBytes[0] = byte(publicKey.E >> 24)
	eBytes[1] = byte(publicKey.E >> 16)
	eBytes[2] = byte(publicKey.E >> 8)
	eBytes[3] = byte(publicKey.E)

	// Убираем ведущие нули
	i := 0
	for i < len(eBytes) && eBytes[i] == 0 {
		i++
	}
	eBytes = eBytes[i:]

	e := base64.RawURLEncoding.EncodeToString(eBytes)

	return dto.JWK{
		KTY: "RSA",
		Use: "sig",
		KID: kid,
		Alg: "RS256",
		N:   n,
		E:   e,
	}, nil
}
