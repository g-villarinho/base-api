package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/g-villarinho/voxel-api/internal/domain"
	"github.com/g-villarinho/voxel-api/internal/infra/database/sqlc"
	"github.com/g-villarinho/voxel-api/internal/mocks"
	"github.com/g-villarinho/voxel-api/pkg/hash"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// securityTestDeps holds all mock dependencies for security service tests.
type securityTestDeps struct {
	store               *mocks.StoreMock
	verificationService *mocks.VerificationServiceMock
	emailService        *mocks.EmailServiceMock
}

func setupSecurityService(t *testing.T) (SecurityService, *securityTestDeps) {
	t.Helper()

	deps := &securityTestDeps{
		store:               mocks.NewStoreMock(t),
		verificationService: mocks.NewVerificationServiceMock(t),
		emailService:        mocks.NewEmailServiceMock(t),
	}

	service := NewSecurityService(
		deps.store,
		deps.verificationService,
		deps.emailService,
	)

	return service, deps
}

func createTestSQLCUserWithPassword(id uuid.UUID, name, email string, status domain.UserStatus, passwordHash string) sqlc.User {
	return sqlc.User{
		ID:           id,
		Name:         name,
		Email:        email,
		Status:       string(status),
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}
}

func createSecurityTestVerification(userID uuid.UUID, flow domain.VerificationFlow, payload *string) *domain.Verification {
	return &domain.Verification{
		ID:        uuid.New(),
		Token:     "test-verification-token",
		UserID:    userID,
		Flow:      flow,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Payload:   payload,
	}
}

func TestUpdatePassword(t *testing.T) {
	t.Parallel()

	t.Run("should return error when user not found", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		userID := uuid.New()
		currentPassword := "currentPassword123"
		newPassword := "newPassword456"
		dbErr := errors.New("user not found")

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(sqlc.User{}, dbErr)

		// Act
		err := service.UpdatePassword(ctx, userID, currentPassword, newPassword)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not find user")
	})

	t.Run("should return ErrPasswordMismatch when user has no password", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		userID := uuid.New()
		currentPassword := "currentPassword123"
		newPassword := "newPassword456"

		userFromDB := sqlc.User{
			ID:           userID,
			Name:         "John Doe",
			Email:        "john@example.com",
			Status:       string(domain.ActiveStatus),
			PasswordHash: "", // No password set
			CreatedAt:    time.Now().UTC(),
		}

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(userFromDB, nil)

		// Act
		err := service.UpdatePassword(ctx, userID, currentPassword, newPassword)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrPasswordMismatch)
	})

	t.Run("should return ErrPasswordMismatch when current password is incorrect", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		userID := uuid.New()
		correctPassword := "correctPassword123"
		wrongPassword := "wrongPassword123"
		newPassword := "newPassword456"

		hashedPassword, err := hash.HashPassword(correctPassword)
		require.NoError(t, err)

		userFromDB := createTestSQLCUserWithPassword(userID, "John Doe", "john@example.com", domain.ActiveStatus, hashedPassword)

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(userFromDB, nil)

		// Act
		err = service.UpdatePassword(ctx, userID, wrongPassword, newPassword)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrPasswordMismatch)
	})

	t.Run("should return error when updating password fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		userID := uuid.New()
		currentPassword := "currentPassword123"
		newPassword := "newPassword456"
		dbErr := errors.New("database error")

		hashedPassword, err := hash.HashPassword(currentPassword)
		require.NoError(t, err)

		userFromDB := createTestSQLCUserWithPassword(userID, "John Doe", "john@example.com", domain.ActiveStatus, hashedPassword)

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(userFromDB, nil)

		deps.store.EXPECT().
			UpdateUserPassword(ctx, mock.AnythingOfType("sqlc.UpdateUserPasswordParams")).
			Return(dbErr)

		// Act
		err = service.UpdatePassword(ctx, userID, currentPassword, newPassword)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not update user password")
	})

	t.Run("should update password successfully", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		userID := uuid.New()
		currentPassword := "currentPassword123"
		newPassword := "newPassword456"

		hashedPassword, err := hash.HashPassword(currentPassword)
		require.NoError(t, err)

		userFromDB := createTestSQLCUserWithPassword(userID, "John Doe", "john@example.com", domain.ActiveStatus, hashedPassword)

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(userFromDB, nil)

		deps.store.EXPECT().
			UpdateUserPassword(ctx, mock.AnythingOfType("sqlc.UpdateUserPasswordParams")).
			Return(nil)

		// Act
		err = service.UpdatePassword(ctx, userID, currentPassword, newPassword)

		// Assert
		require.NoError(t, err)
	})
}

func TestStartPasswordReset(t *testing.T) {
	t.Parallel()

	t.Run("should return nil when user not found to prevent email enumeration", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		email := "nonexistent@example.com"

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(sqlc.User{}, sql.ErrNoRows)

		// Act
		err := service.StartPasswordReset(ctx, email)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should return error when database query fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		email := "john@example.com"
		dbErr := errors.New("database connection error")

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(sqlc.User{}, dbErr)

		// Act
		err := service.StartPasswordReset(ctx, email)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "find user")
	})

	t.Run("should return error when creating verification fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		email := "john@example.com"
		userID := uuid.New()
		verificationErr := errors.New("verification creation error")

		userFromDB := createTestSQLCUserWithPassword(userID, "John Doe", email, domain.ActiveStatus, "hashedpassword")

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(userFromDB, nil)

		deps.verificationService.EXPECT().
			CreateVerification(mock.Anything, mock.AnythingOfType("domain.CreateVerificationParams")).
			Return(nil, verificationErr)

		// Act
		err := service.StartPasswordReset(ctx, email)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create verification")
	})

	t.Run("should start password reset successfully", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		email := "john@example.com"
		userID := uuid.New()

		userFromDB := createTestSQLCUserWithPassword(userID, "John Doe", email, domain.ActiveStatus, "hashedpassword")
		verification := createSecurityTestVerification(userID, domain.ResetPasswordFlow, nil)

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(userFromDB, nil)

		deps.verificationService.EXPECT().
			CreateVerification(mock.Anything, mock.AnythingOfType("domain.CreateVerificationParams")).
			Return(verification, nil)

		deps.emailService.EXPECT().
			SendResetPasswordEmailAsync(ctx, mock.AnythingOfType("*domain.User"), verification).
			Return()

		// Act
		err := service.StartPasswordReset(ctx, email)

		// Assert
		require.NoError(t, err)
	})
}

func TestResetPassword(t *testing.T) {
	t.Parallel()

	t.Run("should return error when token exchange fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		token := "invalid-token"
		newPassword := "newPassword456"
		exchangeErr := errors.New("invalid token")

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.ResetPasswordFlow).
			Return(nil, exchangeErr)

		// Act
		result, err := service.ResetPassword(ctx, token, newPassword)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "consume token")
	})

	t.Run("should return error when updating password fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		token := "valid-token"
		newPassword := "newPassword456"
		userID := uuid.New()
		dbErr := errors.New("database error")

		verification := createSecurityTestVerification(userID, domain.ResetPasswordFlow, nil)

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.ResetPasswordFlow).
			Return(verification, nil)

		deps.store.EXPECT().
			UpdateUserPassword(ctx, mock.AnythingOfType("sqlc.UpdateUserPasswordParams")).
			Return(dbErr)

		// Act
		result, err := service.ResetPassword(ctx, token, newPassword)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "update db password")
	})

	t.Run("should reset password successfully", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		token := "valid-token"
		newPassword := "newPassword456"
		userID := uuid.New()

		verification := createSecurityTestVerification(userID, domain.ResetPasswordFlow, nil)

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.ResetPasswordFlow).
			Return(verification, nil)

		deps.store.EXPECT().
			UpdateUserPassword(ctx, mock.AnythingOfType("sqlc.UpdateUserPasswordParams")).
			Return(nil)

		// Act
		result, err := service.ResetPassword(ctx, token, newPassword)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
	})
}

func TestStartChangeEmail(t *testing.T) {
	t.Parallel()

	t.Run("should return error when user not found", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		userID := uuid.New()
		newEmail := "newemail@example.com"
		dbErr := errors.New("user not found")

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(sqlc.User{}, dbErr)

		// Act
		err := service.StartChangeEmail(ctx, userID, newEmail)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "find user by id")
	})

	t.Run("should return ErrEmailIsTheSame when new email equals current email", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		userID := uuid.New()
		email := "john@example.com"

		userFromDB := createTestSQLCUserWithPassword(userID, "John Doe", email, domain.ActiveStatus, "hashedpassword")

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(userFromDB, nil)

		// Act
		err := service.StartChangeEmail(ctx, userID, email)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrEmailIsTheSame)
	})

	t.Run("should return error when checking email existence fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		userID := uuid.New()
		currentEmail := "john@example.com"
		newEmail := "newemail@example.com"
		dbErr := errors.New("database error")

		userFromDB := createTestSQLCUserWithPassword(userID, "John Doe", currentEmail, domain.ActiveStatus, "hashedpassword")

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(userFromDB, nil)

		deps.store.EXPECT().
			ExistsByEmail(ctx, newEmail).
			Return(false, dbErr)

		// Act
		err := service.StartChangeEmail(ctx, userID, newEmail)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "check email existence")
	})

	t.Run("should return ErrEmailAlreadyExists when new email is already in use", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		userID := uuid.New()
		currentEmail := "john@example.com"
		newEmail := "existing@example.com"

		userFromDB := createTestSQLCUserWithPassword(userID, "John Doe", currentEmail, domain.ActiveStatus, "hashedpassword")

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(userFromDB, nil)

		deps.store.EXPECT().
			ExistsByEmail(ctx, newEmail).
			Return(true, nil)

		// Act
		err := service.StartChangeEmail(ctx, userID, newEmail)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
	})

	t.Run("should return error when creating verification fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		userID := uuid.New()
		currentEmail := "john@example.com"
		newEmail := "newemail@example.com"
		verificationErr := errors.New("verification creation error")

		userFromDB := createTestSQLCUserWithPassword(userID, "John Doe", currentEmail, domain.ActiveStatus, "hashedpassword")

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(userFromDB, nil)

		deps.store.EXPECT().
			ExistsByEmail(ctx, newEmail).
			Return(false, nil)

		deps.verificationService.EXPECT().
			CreateVerification(mock.Anything, mock.AnythingOfType("domain.CreateVerificationParams")).
			Return(nil, verificationErr)

		// Act
		err := service.StartChangeEmail(ctx, userID, newEmail)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create verification")
	})

	t.Run("should start email change successfully", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		userID := uuid.New()
		currentEmail := "john@example.com"
		newEmail := "newemail@example.com"

		userFromDB := createTestSQLCUserWithPassword(userID, "John Doe", currentEmail, domain.ActiveStatus, "hashedpassword")
		verification := createSecurityTestVerification(userID, domain.ChangeEmailFlow, &newEmail)

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(userFromDB, nil)

		deps.store.EXPECT().
			ExistsByEmail(ctx, newEmail).
			Return(false, nil)

		deps.verificationService.EXPECT().
			CreateVerification(mock.Anything, mock.AnythingOfType("domain.CreateVerificationParams")).
			Return(verification, nil)

		deps.emailService.EXPECT().
			SendChangeEmailAsync(ctx, mock.AnythingOfType("*domain.User"), verification).
			Return()

		// Act
		err := service.StartChangeEmail(ctx, userID, newEmail)

		// Assert
		require.NoError(t, err)
	})
}

func TestChangeEmail(t *testing.T) {
	t.Parallel()

	t.Run("should return error when token exchange fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		token := "invalid-token"
		exchangeErr := errors.New("invalid token")

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.ChangeEmailFlow).
			Return(nil, exchangeErr)

		// Act
		err := service.ChangeEmail(ctx, token)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "consume token")
	})

	t.Run("should return ErrInvalidVerificationPayload when payload is nil", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		token := "valid-token"
		userID := uuid.New()

		verification := createSecurityTestVerification(userID, domain.ChangeEmailFlow, nil)

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.ChangeEmailFlow).
			Return(verification, nil)

		// Act
		err := service.ChangeEmail(ctx, token)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidVerificationPayload)
	})

	t.Run("should return error when updating email fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		token := "valid-token"
		userID := uuid.New()
		newEmail := "newemail@example.com"
		dbErr := errors.New("database error")

		verification := createSecurityTestVerification(userID, domain.ChangeEmailFlow, &newEmail)

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.ChangeEmailFlow).
			Return(verification, nil)

		deps.store.EXPECT().
			UpdateUserEmail(ctx, mock.AnythingOfType("sqlc.UpdateUserEmailParams")).
			Return(dbErr)

		// Act
		err := service.ChangeEmail(ctx, token)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "update db email")
	})

	t.Run("should change email successfully", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSecurityService(t)
		ctx := context.Background()
		token := "valid-token"
		userID := uuid.New()
		newEmail := "newemail@example.com"

		verification := createSecurityTestVerification(userID, domain.ChangeEmailFlow, &newEmail)

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.ChangeEmailFlow).
			Return(verification, nil)

		deps.store.EXPECT().
			UpdateUserEmail(ctx, mock.AnythingOfType("sqlc.UpdateUserEmailParams")).
			Return(nil)

		// Act
		err := service.ChangeEmail(ctx, token)

		// Assert
		require.NoError(t, err)
	})
}
