package crypto

import (
	"crypto/subtle"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword  = errors.New("invalid password")
	ErrPasswordTooShort = errors.New("password is too short")
)

// PasswordHasher - интерфейс для хеширования паролей
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) error
}

// BcryptHasher - bcrypt реализация
type BcryptHasher struct {
	cost      int
	minLength int
}

func NewBcryptHasher(cost, minLength int) *BcryptHasher {
	if cost < bcrypt.MinCost {
		cost = bcrypt.DefaultCost
	}
	if minLength < 8 {
		minLength = 8
	}
	return &BcryptHasher{
		cost:      cost,
		minLength: minLength,
	}
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	if len(password) < h.minLength {
		return "", ErrPasswordTooShort
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *BcryptHasher) Verify(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return err
	}

	// Constant-time сравнение для защиты от timing attacks
	// (bcrypt уже делает это внутри, но добавим для явности)
	if subtle.ConstantTimeCompare([]byte(hash), []byte(hash)) != 1 {
		return ErrInvalidPassword
	}

	return nil
}
