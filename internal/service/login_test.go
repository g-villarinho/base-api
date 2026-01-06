package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/g-villarinho/voxel-api/config"
	"github.com/g-villarinho/voxel-api/internal/domain"
	"github.com/g-villarinho/voxel-api/internal/infra/database/sqlc"
	"github.com/g-villarinho/voxel-api/internal/mocks"
	"github.com/g-villarinho/voxel-api/pkg/hash"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// loginTestDeps holds all mock dependencies for login service tests.
type loginTestDeps struct {
	store               *mocks.StoreMock
	verificationService *mocks.VerificationServiceMock
	emailService        *mocks.EmailServiceMock
	config              *config.Config
}

func setupLoginService(t *testing.T) (LoginService, *loginTestDeps) {
	t.Helper()

	deps := &loginTestDeps{
		store:               mocks.NewStoreMock(t),
		verificationService: mocks.NewVerificationServiceMock(t),
		emailService:        mocks.NewEmailServiceMock(t),
		config: &config.Config{
			Auth: config.Auth{
				Method: config.AuthMethodBoth,
			},
		},
	}

	service := NewLoginService(
		deps.store,
		deps.verificationService,
		deps.emailService,
		deps.config,
	)

	return service, deps
}

func createTestSQLCUser(id uuid.UUID, name, email string, status domain.UserStatus, passwordHash string, emailConfirmedAt *time.Time) sqlc.User {
	return sqlc.User{
		ID:               id,
		Name:             name,
		Email:            email,
		Status:           string(status),
		PasswordHash:     passwordHash,
		CreatedAt:        time.Now().UTC(),
		EmailConfirmedAt: emailConfirmedAt,
	}
}

func TestAuthenticate(t *testing.T) {
	t.Parallel()

	t.Run("should return error when password auth is not enabled", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		deps.config.Auth.Method = config.AuthMethodMagicLink
		ctx := context.Background()
		email := "john@example.com"
		password := "password123"

		// Act
		user, err := service.Authenticate(ctx, email, password)

		// Assert
		require.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrPasswordAuthNotEnabled)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		email := "nonexistent@example.com"
		password := "password123"

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(sqlc.User{}, sql.ErrNoRows)

		// Act
		user, err := service.Authenticate(ctx, email, password)

		// Assert
		require.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("should return error when database fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		email := "john@example.com"
		password := "password123"
		dbErr := errors.New("database connection error")

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(sqlc.User{}, dbErr)

		// Act
		user, err := service.Authenticate(ctx, email, password)

		// Assert
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "find user by email")
	})

	t.Run("should return error when user has no password", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		email := "john@example.com"
		password := "password123"
		userID := uuid.New()
		emailConfirmedAt := time.Now().UTC()

		sqlcUser := createTestSQLCUser(userID, "John Doe", email, domain.ActiveStatus, "", &emailConfirmedAt)

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(sqlcUser, nil)

		// Act
		user, err := service.Authenticate(ctx, email, password)

		// Assert
		require.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("should return error when password is invalid", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		email := "john@example.com"
		password := "wrongpassword"
		userID := uuid.New()
		emailConfirmedAt := time.Now().UTC()

		hashedPassword, _ := hash.HashPassword("correctpassword")
		sqlcUser := createTestSQLCUser(userID, "John Doe", email, domain.ActiveStatus, hashedPassword, &emailConfirmedAt)

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(sqlcUser, nil)

		// Act
		user, err := service.Authenticate(ctx, email, password)

		// Assert
		require.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("should return error when user is blocked", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		email := "blocked@example.com"
		password := "password123"
		userID := uuid.New()
		emailConfirmedAt := time.Now().UTC()

		hashedPassword, _ := hash.HashPassword(password)
		sqlcUser := createTestSQLCUser(userID, "Blocked User", email, domain.BlockedStatus, hashedPassword, &emailConfirmedAt)

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(sqlcUser, nil)

		// Act
		user, err := service.Authenticate(ctx, email, password)

		// Assert
		require.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrUserBlocked)
	})

	t.Run("should return error when email is not verified", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		email := "unverified@example.com"
		password := "password123"
		userID := uuid.New()

		hashedPassword, _ := hash.HashPassword(password)
		sqlcUser := createTestSQLCUser(userID, "Unverified User", email, domain.ActiveStatus, hashedPassword, nil)

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(sqlcUser, nil)

		// Act
		user, err := service.Authenticate(ctx, email, password)

		// Assert
		require.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrEmailNotVerified)
	})

	t.Run("should return user when credentials are valid", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		email := "john@example.com"
		password := "password123"
		userID := uuid.New()
		emailConfirmedAt := time.Now().UTC()

		hashedPassword, _ := hash.HashPassword(password)
		sqlcUser := createTestSQLCUser(userID, "John Doe", email, domain.ActiveStatus, hashedPassword, &emailConfirmedAt)

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(sqlcUser, nil)

		// Act
		user, err := service.Authenticate(ctx, email, password)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, "John Doe", user.Name)
	})
}

func TestRequestMagicLink(t *testing.T) {
	t.Parallel()

	t.Run("should return error when magic link auth is not enabled", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		deps.config.Auth.Method = config.AuthMethodPassword
		ctx := context.Background()
		email := "john@example.com"

		// Act
		err := service.RequestMagicLink(ctx, email)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrMagicLinkNotEnabled)
	})

	t.Run("should return nil when user not found to avoid revealing email existence", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		email := "nonexistent@example.com"

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(sqlc.User{}, sql.ErrNoRows)

		// Act
		err := service.RequestMagicLink(ctx, email)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should return error when database fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		email := "john@example.com"
		dbErr := errors.New("database connection error")

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(sqlc.User{}, dbErr)

		// Act
		err := service.RequestMagicLink(ctx, email)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "find user")
	})

	t.Run("should return nil when user is blocked to avoid revealing blocked status", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		email := "blocked@example.com"
		userID := uuid.New()

		sqlcUser := createTestSQLCUser(userID, "Blocked User", email, domain.BlockedStatus, "", nil)

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(sqlcUser, nil)

		// Act
		err := service.RequestMagicLink(ctx, email)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should return error when verification creation fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		email := "john@example.com"
		userID := uuid.New()
		verificationErr := errors.New("verification creation failed")

		sqlcUser := createTestSQLCUser(userID, "John Doe", email, domain.ActiveStatus, "", nil)

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(sqlcUser, nil)

		deps.verificationService.EXPECT().
			CreateVerification(mock.Anything, mock.AnythingOfType("domain.CreateVerificationParams")).
			Return(nil, verificationErr)

		// Act
		err := service.RequestMagicLink(ctx, email)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create magic link")
	})

	t.Run("should send magic link email successfully", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		email := "john@example.com"
		userID := uuid.New()

		sqlcUser := createTestSQLCUser(userID, "John Doe", email, domain.ActiveStatus, "", nil)
		verification := createTestVerification(userID, domain.MagicLinkLoginFlow)

		deps.store.EXPECT().
			FindUserByEmail(ctx, email).
			Return(sqlcUser, nil)

		deps.verificationService.EXPECT().
			CreateVerification(mock.Anything, mock.AnythingOfType("domain.CreateVerificationParams")).
			Return(verification, nil)

		deps.emailService.EXPECT().
			SendMagicLinkEmailAsync(ctx, mock.AnythingOfType("*domain.User"), verification).
			Return()

		// Act
		err := service.RequestMagicLink(ctx, email)

		// Assert
		require.NoError(t, err)
	})
}

func TestVerifyMagicLink(t *testing.T) {
	t.Parallel()

	t.Run("should return error when token exchange fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		token := "invalid-token"
		tokenErr := errors.New("token is invalid or expired")

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.MagicLinkLoginFlow).
			Return(nil, tokenErr)

		// Act
		user, err := service.VerifyMagicLink(ctx, token)

		// Assert
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "invalid token")
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		token := "valid-token"
		userID := uuid.New()
		verification := createTestVerification(userID, domain.MagicLinkLoginFlow)

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.MagicLinkLoginFlow).
			Return(verification, nil)

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(sqlc.User{}, sql.ErrNoRows)

		// Act
		user, err := service.VerifyMagicLink(ctx, token)

		// Assert
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "find user by id")
	})

	t.Run("should return error when user is blocked", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		token := "valid-token"
		userID := uuid.New()
		verification := createTestVerification(userID, domain.MagicLinkLoginFlow)

		sqlcUser := createTestSQLCUser(userID, "Blocked User", "blocked@example.com", domain.BlockedStatus, "", nil)

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.MagicLinkLoginFlow).
			Return(verification, nil)

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(sqlcUser, nil)

		// Act
		user, err := service.VerifyMagicLink(ctx, token)

		// Assert
		require.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrUserBlocked)
	})

	t.Run("should verify email when email is not yet verified", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		token := "valid-token"
		userID := uuid.New()
		verification := createTestVerification(userID, domain.MagicLinkLoginFlow)

		sqlcUser := createTestSQLCUser(userID, "John Doe", "john@example.com", domain.ActiveStatus, "", nil)

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.MagicLinkLoginFlow).
			Return(verification, nil)

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(sqlcUser, nil)

		deps.store.EXPECT().
			VerifyUserEmail(ctx, mock.AnythingOfType("sqlc.VerifyUserEmailParams")).
			Return(nil)

		// Act
		user, err := service.VerifyMagicLink(ctx, token)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
	})

	t.Run("should return error when email verification update fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		token := "valid-token"
		userID := uuid.New()
		verification := createTestVerification(userID, domain.MagicLinkLoginFlow)
		updateErr := errors.New("update failed")

		sqlcUser := createTestSQLCUser(userID, "John Doe", "john@example.com", domain.ActiveStatus, "", nil)

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.MagicLinkLoginFlow).
			Return(verification, nil)

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(sqlcUser, nil)

		deps.store.EXPECT().
			VerifyUserEmail(ctx, mock.AnythingOfType("sqlc.VerifyUserEmailParams")).
			Return(updateErr)

		// Act
		user, err := service.VerifyMagicLink(ctx, token)

		// Assert
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "verify user email")
	})

	t.Run("should return user without verifying email when email is already verified", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupLoginService(t)
		ctx := context.Background()
		token := "valid-token"
		userID := uuid.New()
		verification := createTestVerification(userID, domain.MagicLinkLoginFlow)
		emailConfirmedAt := time.Now().UTC()

		sqlcUser := createTestSQLCUser(userID, "John Doe", "john@example.com", domain.ActiveStatus, "", &emailConfirmedAt)

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.MagicLinkLoginFlow).
			Return(verification, nil)

		deps.store.EXPECT().
			FindUserByID(ctx, userID).
			Return(sqlcUser, nil)

		// Act
		user, err := service.VerifyMagicLink(ctx, token)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "john@example.com", user.Email)
	})
}
