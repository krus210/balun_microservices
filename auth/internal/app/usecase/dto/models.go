package dto

type RegisterRequest struct {
	Email    string
	Password string
}

type LoginRequest struct {
	Email    string
	Password string
	DeviceID string
}

type RefreshRequest struct {
	RefreshToken string
	DeviceID     string
}

type LogoutRequest struct {
	RefreshToken string
}

// JWKSResponse - формат ответа JWKS endpoint
type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

// JWK - JSON Web Key в формате RSA
type JWK struct {
	KTY string `json:"kty"` // Key Type (RSA)
	Use string `json:"use"` // Public Key Use (sig)
	KID string `json:"kid"` // Key ID
	Alg string `json:"alg"` // Algorithm (RS256)
	N   string `json:"n"`   // Modulus (base64url)
	E   string `json:"e"`   // Exponent (base64url)
}
