package dto

type RegisterRequest struct {
	Email    string
	Password string
}

type LoginRequest struct {
	Email    string
	Password string
}

type RefreshRequest struct {
	UserID       string
	RefreshToken string
}
