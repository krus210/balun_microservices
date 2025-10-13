package models

import "errors"

var (
	// Ошибки поиска
	ErrNotFound = errors.New("profile not found")

	// Ошибки создания/обновления
	ErrAlreadyExists = errors.New("user profile already exists")
)
