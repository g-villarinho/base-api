package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/gbvillarinho/base-project/internal/domain"
	"github.com/gbvillarinho/base-project/internal/infra/database/sqlc"
	"github.com/gbvillarinho/base-project/internal/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserService(t *testing.T) (UserService, *mocks.StoreMock) {
	t.Helper()

	store := mocks.NewStoreMock(t)
	service := NewUserService(store)

	return service, store
}

func TestGetUser(t *testing.T) {
	t.Parallel()

	t.Run("should return user when user exists", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, store := setupUserService(t)
		ctx := context.Background()
		userID := uuid.New()
		emailConfirmedAt := time.Now().UTC()

		sqlcUser := sqlc.User{
			ID:               userID,
			Name:             "John Doe",
			Email:            "john@example.com",
			Status:           string(domain.ActiveStatus),
			PasswordHash:     "hashed-password",
			CreatedAt:        time.Now().UTC(),
			EmailConfirmedAt: &emailConfirmedAt,
		}

		store.EXPECT().
			FindUserByID(ctx, userID).
			Return(sqlcUser, nil)

		// Act
		user, err := service.GetUser(ctx, userID)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "john@example.com", user.Email)
		assert.Equal(t, domain.ActiveStatus, user.Status)
	})

	t.Run("should return ErrUserNotFound when user does not exist", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, store := setupUserService(t)
		ctx := context.Background()
		userID := uuid.New()

		store.EXPECT().
			FindUserByID(ctx, userID).
			Return(sqlc.User{}, sql.ErrNoRows)

		// Act
		user, err := service.GetUser(ctx, userID)

		// Assert
		require.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("should return error when database fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, store := setupUserService(t)
		ctx := context.Background()
		userID := uuid.New()
		dbErr := errors.New("database connection error")

		store.EXPECT().
			FindUserByID(ctx, userID).
			Return(sqlc.User{}, dbErr)

		// Act
		user, err := service.GetUser(ctx, userID)

		// Assert
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "find user by id")
	})

	t.Run("should return user without password when user has no password", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, store := setupUserService(t)
		ctx := context.Background()
		userID := uuid.New()

		sqlcUser := sqlc.User{
			ID:           userID,
			Name:         "Jane Doe",
			Email:        "jane@example.com",
			Status:       string(domain.ActiveStatus),
			PasswordHash: "",
			CreatedAt:    time.Now().UTC(),
		}

		store.EXPECT().
			FindUserByID(ctx, userID).
			Return(sqlcUser, nil)

		// Act
		user, err := service.GetUser(ctx, userID)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "Jane Doe", user.Name)
		assert.Nil(t, user.PasswordHash)
	})

	t.Run("should return user with blocked status", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, store := setupUserService(t)
		ctx := context.Background()
		userID := uuid.New()
		blockedAt := time.Now().UTC()

		sqlcUser := sqlc.User{
			ID:        userID,
			Name:      "Blocked User",
			Email:     "blocked@example.com",
			Status:    string(domain.BlockedStatus),
			CreatedAt: time.Now().UTC(),
			BlockedAt: &blockedAt,
		}

		store.EXPECT().
			FindUserByID(ctx, userID).
			Return(sqlcUser, nil)

		// Act
		user, err := service.GetUser(ctx, userID)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, domain.BlockedStatus, user.Status)
		assert.NotNil(t, user.BlockedAt)
	})
}
