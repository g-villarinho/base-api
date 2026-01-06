package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gbvillarinho/base-project/internal/domain"
	"github.com/gbvillarinho/base-project/internal/infra/database/sqlc"
	"github.com/google/uuid"
)

type UserService interface {
	GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error)
}

type userService struct {
	store sqlc.Store
}

func NewUserService(store sqlc.Store) UserService {
	return &userService{
		store: store,
	}
}

func (s *userService) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	userFromDB, err := s.store.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return toDomainUser(userFromDB), nil
}
