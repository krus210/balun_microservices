package models

import "errors"

var (
	// Ошибки поиска
	ErrNotFound = errors.New("profile not found")

	// Ошибки создания/обновления
	ErrNicknameAlreadyExists = errors.New("nickname already exists")
	ErrUserAlreadyExists     = errors.New("user profile already exists")
)
