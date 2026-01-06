package domain

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
)

type LoginResult struct {
	UserID uuid.UUID
}

type RegisterResult struct {
	UserID uuid.UUID
}

type ResetPasswordResult struct {
	UserID uuid.UUID
}

type ConfirmEmailResult struct {
	UserID uuid.UUID
}
