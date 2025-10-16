package models

import "errors"

var (
	ErrNotFound         = errors.New("chat not found")
	ErrAlreadyExists    = errors.New("chat already exists")
	ErrPermissionDenied = errors.New("permission denied")
)
