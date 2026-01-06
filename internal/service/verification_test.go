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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupVerificationService(t *testing.T) (VerificationService, *mocks.StoreMock) {
	t.Helper()

	storeMock := mocks.NewStoreMock(t)
	service := NewVerificationService(storeMock)

	return service, storeMock
}

func createTestSQLCVerification(id, userID uuid.UUID, flow string, expiresAt time.Time, payload *string) sqlc.Verification {
	return sqlc.Verification{
		ID:        id,
		Flow:      flow,
		Token:     "test-token-12345",
		CreatedAt: time.Now().UTC(),
		ExpiresAt: expiresAt,
		Payload:   payload,
		UserID:    userID,
	}
}

func TestCreateVerification(t *testing.T) {
	t.Parallel()

	t.Run("should return existing verification when valid verification exists with more than 5 minutes remaining", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, storeMock := setupVerificationService(t)
		ctx := context.Background()
		userID := uuid.New()
		verificationID := uuid.New()

		params := domain.CreateVerificationParams{
			UserID:  userID,
			Flow:    domain.VerificationEmailFlow,
			Payload: "",
		}

		expiresAt := time.Now().UTC().Add(8 * time.Minute)
		existingVerification := createTestSQLCVerification(verificationID, userID, string(domain.VerificationEmailFlow), expiresAt, nil)

		storeMock.EXPECT().
			FindValidVerificationByUserIDAndFlow(ctx, mock.AnythingOfType("sqlc.FindValidVerificationByUserIDAndFlowParams")).
			Return(existingVerification, nil)

		// Act
		verification, err := service.CreateVerification(ctx, params)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, verification)
		assert.Equal(t, verificationID, verification.ID)
		assert.Equal(t, userID, verification.UserID)
		assert.Equal(t, domain.VerificationEmailFlow, verification.Flow)
	})

	t.Run("should create new verification when existing verification has less than 5 minutes remaining", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, storeMock := setupVerificationService(t)
		ctx := context.Background()
		userID := uuid.New()

		params := domain.CreateVerificationParams{
			UserID:  userID,
			Flow:    domain.VerificationEmailFlow,
			Payload: "",
		}

		expiresAt := time.Now().UTC().Add(3 * time.Minute)
		existingVerification := createTestSQLCVerification(uuid.New(), userID, string(domain.VerificationEmailFlow), expiresAt, nil)

		storeMock.EXPECT().
			FindValidVerificationByUserIDAndFlow(ctx, mock.AnythingOfType("sqlc.FindValidVerificationByUserIDAndFlowParams")).
			Return(existingVerification, nil)

		storeMock.EXPECT().
			CreateVerification(ctx, mock.AnythingOfType("sqlc.CreateVerificationParams")).
			Return(nil)

		// Act
		verification, err := service.CreateVerification(ctx, params)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, verification)
		assert.Equal(t, userID, verification.UserID)
		assert.Equal(t, domain.VerificationEmailFlow, verification.Flow)
	})

	t.Run("should create new verification when no existing verification found", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, storeMock := setupVerificationService(t)
		ctx := context.Background()
		userID := uuid.New()

		params := domain.CreateVerificationParams{
			UserID:  userID,
			Flow:    domain.ResetPasswordFlow,
			Payload: "",
		}

		storeMock.EXPECT().
			FindValidVerificationByUserIDAndFlow(ctx, mock.AnythingOfType("sqlc.FindValidVerificationByUserIDAndFlowParams")).
			Return(sqlc.Verification{}, sql.ErrNoRows)

		storeMock.EXPECT().
			CreateVerification(ctx, mock.AnythingOfType("sqlc.CreateVerificationParams")).
			Return(nil)

		// Act
		verification, err := service.CreateVerification(ctx, params)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, verification)
		assert.Equal(t, userID, verification.UserID)
		assert.Equal(t, domain.ResetPasswordFlow, verification.Flow)
		assert.NotEmpty(t, verification.Token)
	})

	t.Run("should create new verification with payload", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, storeMock := setupVerificationService(t)
		ctx := context.Background()
		userID := uuid.New()
		payload := "new-email@example.com"

		params := domain.CreateVerificationParams{
			UserID:  userID,
			Flow:    domain.ChangeEmailFlow,
			Payload: payload,
		}

		storeMock.EXPECT().
			FindValidVerificationByUserIDAndFlow(ctx, mock.AnythingOfType("sqlc.FindValidVerificationByUserIDAndFlowParams")).
			Return(sqlc.Verification{}, sql.ErrNoRows)

		storeMock.EXPECT().
			CreateVerification(ctx, mock.AnythingOfType("sqlc.CreateVerificationParams")).
			Return(nil)

		// Act
		verification, err := service.CreateVerification(ctx, params)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, verification)
		assert.NotNil(t, verification.Payload)
		assert.Equal(t, payload, *verification.Payload)
	})

	t.Run("should return error when finding existing verification fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, storeMock := setupVerificationService(t)
		ctx := context.Background()
		userID := uuid.New()
		dbErr := errors.New("database connection error")

		params := domain.CreateVerificationParams{
			UserID:  userID,
			Flow:    domain.VerificationEmailFlow,
			Payload: "",
		}

		storeMock.EXPECT().
			FindValidVerificationByUserIDAndFlow(ctx, mock.AnythingOfType("sqlc.FindValidVerificationByUserIDAndFlowParams")).
			Return(sqlc.Verification{}, dbErr)

		// Act
		verification, err := service.CreateVerification(ctx, params)

		// Assert
		require.Error(t, err)
		assert.Nil(t, verification)
		assert.Contains(t, err.Error(), "find valid verification")
	})

	t.Run("should return error when creating verification fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, storeMock := setupVerificationService(t)
		ctx := context.Background()
		userID := uuid.New()
		createErr := errors.New("failed to create verification")

		params := domain.CreateVerificationParams{
			UserID:  userID,
			Flow:    domain.MagicLinkLoginFlow,
			Payload: "",
		}

		storeMock.EXPECT().
			FindValidVerificationByUserIDAndFlow(ctx, mock.AnythingOfType("sqlc.FindValidVerificationByUserIDAndFlowParams")).
			Return(sqlc.Verification{}, sql.ErrNoRows)

		storeMock.EXPECT().
			CreateVerification(ctx, mock.AnythingOfType("sqlc.CreateVerificationParams")).
			Return(createErr)

		// Act
		verification, err := service.CreateVerification(ctx, params)

		// Assert
		require.Error(t, err)
		assert.Nil(t, verification)
		assert.Contains(t, err.Error(), "create verification")
	})
}

func TestExchangeToken(t *testing.T) {
	t.Parallel()

	t.Run("should return verification when token is valid and not expired", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, storeMock := setupVerificationService(t)
		ctx := context.Background()
		token := "valid-token-12345"
		userID := uuid.New()
		verificationID := uuid.New()

		expiresAt := time.Now().UTC().Add(5 * time.Minute)
		sqlcVerification := createTestSQLCVerification(verificationID, userID, string(domain.VerificationEmailFlow), expiresAt, nil)
		sqlcVerification.Token = token

		storeMock.EXPECT().
			FindVerificationByToken(ctx, token).
			Return(sqlcVerification, nil)

		storeMock.EXPECT().
			DeleteVerification(ctx, verificationID).
			Return(nil)

		// Act
		verification, err := service.ExchangeToken(ctx, token, domain.VerificationEmailFlow)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, verification)
		assert.Equal(t, verificationID, verification.ID)
		assert.Equal(t, userID, verification.UserID)
		assert.Equal(t, domain.VerificationEmailFlow, verification.Flow)
	})

	t.Run("should return error when token not found", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, storeMock := setupVerificationService(t)
		ctx := context.Background()
		token := "nonexistent-token"

		storeMock.EXPECT().
			FindVerificationByToken(ctx, token).
			Return(sqlc.Verification{}, sql.ErrNoRows)

		// Act
		verification, err := service.ExchangeToken(ctx, token, domain.VerificationEmailFlow)

		// Assert
		require.Error(t, err)
		assert.Nil(t, verification)
		assert.ErrorIs(t, err, domain.ErrInvalidVerification)
	})

	t.Run("should return error when database fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, storeMock := setupVerificationService(t)
		ctx := context.Background()
		token := "valid-token"
		dbErr := errors.New("database connection error")

		storeMock.EXPECT().
			FindVerificationByToken(ctx, token).
			Return(sqlc.Verification{}, dbErr)

		// Act
		verification, err := service.ExchangeToken(ctx, token, domain.VerificationEmailFlow)

		// Assert
		require.Error(t, err)
		assert.Nil(t, verification)
		assert.Contains(t, err.Error(), "find verification by token")
	})

	t.Run("should return error when verification is expired", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, storeMock := setupVerificationService(t)
		ctx := context.Background()
		token := "expired-token"
		userID := uuid.New()
		verificationID := uuid.New()

		expiresAt := time.Now().UTC().Add(-5 * time.Minute)
		sqlcVerification := createTestSQLCVerification(verificationID, userID, string(domain.VerificationEmailFlow), expiresAt, nil)
		sqlcVerification.Token = token

		storeMock.EXPECT().
			FindVerificationByToken(ctx, token).
			Return(sqlcVerification, nil)

		// Act
		verification, err := service.ExchangeToken(ctx, token, domain.VerificationEmailFlow)

		// Assert
		require.Error(t, err)
		assert.Nil(t, verification)
		assert.ErrorIs(t, err, domain.ErrInvalidVerification)
	})

	t.Run("should return error when flow does not match expected flow", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, storeMock := setupVerificationService(t)
		ctx := context.Background()
		token := "valid-token"
		userID := uuid.New()
		verificationID := uuid.New()

		expiresAt := time.Now().UTC().Add(5 * time.Minute)
		sqlcVerification := createTestSQLCVerification(verificationID, userID, string(domain.ResetPasswordFlow), expiresAt, nil)
		sqlcVerification.Token = token

		storeMock.EXPECT().
			FindVerificationByToken(ctx, token).
			Return(sqlcVerification, nil)

		// Act
		verification, err := service.ExchangeToken(ctx, token, domain.VerificationEmailFlow)

		// Assert
		require.Error(t, err)
		assert.Nil(t, verification)
		assert.ErrorIs(t, err, domain.ErrInvalidVerification)
	})

	t.Run("should return error when deleting verification fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, storeMock := setupVerificationService(t)
		ctx := context.Background()
		token := "valid-token"
		userID := uuid.New()
		verificationID := uuid.New()
		deleteErr := errors.New("failed to delete verification")

		expiresAt := time.Now().UTC().Add(5 * time.Minute)
		sqlcVerification := createTestSQLCVerification(verificationID, userID, string(domain.MagicLinkLoginFlow), expiresAt, nil)
		sqlcVerification.Token = token

		storeMock.EXPECT().
			FindVerificationByToken(ctx, token).
			Return(sqlcVerification, nil)

		storeMock.EXPECT().
			DeleteVerification(ctx, verificationID).
			Return(deleteErr)

		// Act
		verification, err := service.ExchangeToken(ctx, token, domain.MagicLinkLoginFlow)

		// Assert
		require.Error(t, err)
		assert.Nil(t, verification)
		assert.Contains(t, err.Error(), "delete verification")
	})

	t.Run("should return verification with payload when token is valid", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, storeMock := setupVerificationService(t)
		ctx := context.Background()
		token := "valid-token"
		userID := uuid.New()
		verificationID := uuid.New()
		payload := "new-email@example.com"

		expiresAt := time.Now().UTC().Add(5 * time.Minute)
		sqlcVerification := createTestSQLCVerification(verificationID, userID, string(domain.ChangeEmailFlow), expiresAt, &payload)
		sqlcVerification.Token = token

		storeMock.EXPECT().
			FindVerificationByToken(ctx, token).
			Return(sqlcVerification, nil)

		storeMock.EXPECT().
			DeleteVerification(ctx, verificationID).
			Return(nil)

		// Act
		verification, err := service.ExchangeToken(ctx, token, domain.ChangeEmailFlow)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, verification)
		assert.NotNil(t, verification.Payload)
		assert.Equal(t, payload, *verification.Payload)
	})
}
