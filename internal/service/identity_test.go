package service

import (
	"context"
	"errors"
	"testing"

	"github.com/g-villarinho/voxel-api/config"
	"github.com/g-villarinho/voxel-api/internal/domain"
	"github.com/g-villarinho/voxel-api/internal/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// identityTestDeps holds all mock dependencies for identity service tests.
type identityTestDeps struct {
	store               *mocks.StoreMock
	verificationService *mocks.VerificationServiceMock
	emailService        *mocks.EmailServiceMock
}

func setupIdentityService(t *testing.T, authMethod string) (IdentityService, *identityTestDeps) {
	t.Helper()

	deps := &identityTestDeps{
		store:               mocks.NewStoreMock(t),
		verificationService: mocks.NewVerificationServiceMock(t),
		emailService:        mocks.NewEmailServiceMock(t),
	}

	cfg := &config.Config{
		Auth: config.Auth{
			Method: authMethod,
		},
	}

	service := NewIdentityService(
		deps.store,
		deps.verificationService,
		deps.emailService,
		cfg,
	)

	return service, deps
}

func createTestVerification(userID uuid.UUID, flow domain.VerificationFlow) *domain.Verification {
	return &domain.Verification{
		ID:     uuid.New(),
		Token:  "verification-token",
		UserID: userID,
		Flow:   flow,
	}
}

func TestRegister(t *testing.T) {
	t.Parallel()

	t.Run("should return ErrPasswordAuthNotEnabled when password auth is disabled", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, _ := setupIdentityService(t, config.AuthMethodMagicLink)
		ctx := context.Background()
		name := "John Doe"
		email := "john@example.com"
		password := "securePassword123"

		// Act
		err := service.Register(ctx, name, email, password)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrPasswordAuthNotEnabled)
	})

	t.Run("should return error when ExistsByEmail fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupIdentityService(t, config.AuthMethodPassword)
		ctx := context.Background()
		name := "John Doe"
		email := "john@example.com"
		password := "securePassword123"
		dbErr := errors.New("database connection failed")

		deps.store.EXPECT().
			ExistsByEmail(ctx, email).
			Return(false, dbErr)

		// Act
		err := service.Register(ctx, name, email, password)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "email already exists")
	})

	t.Run("should return ErrEmailAlreadyExists when email already exists", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupIdentityService(t, config.AuthMethodPassword)
		ctx := context.Background()
		name := "John Doe"
		email := "john@example.com"
		password := "securePassword123"

		deps.store.EXPECT().
			ExistsByEmail(ctx, email).
			Return(true, nil)

		// Act
		err := service.Register(ctx, name, email, password)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
	})

	t.Run("should return error when CreateUser fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupIdentityService(t, config.AuthMethodPassword)
		ctx := context.Background()
		name := "John Doe"
		email := "john@example.com"
		password := "securePassword123"
		dbErr := errors.New("failed to create user")

		deps.store.EXPECT().
			ExistsByEmail(ctx, email).
			Return(false, nil)

		deps.store.EXPECT().
			CreateUser(ctx, mock.AnythingOfType("sqlc.CreateUserParams")).
			Return(dbErr)

		// Act
		err := service.Register(ctx, name, email, password)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create user")
	})

	t.Run("should return error when CreateVerification fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupIdentityService(t, config.AuthMethodPassword)
		ctx := context.Background()
		name := "John Doe"
		email := "john@example.com"
		password := "securePassword123"
		verificationErr := errors.New("failed to create verification")

		deps.store.EXPECT().
			ExistsByEmail(ctx, email).
			Return(false, nil)

		deps.store.EXPECT().
			CreateUser(ctx, mock.AnythingOfType("sqlc.CreateUserParams")).
			Return(nil)

		deps.verificationService.EXPECT().
			CreateVerification(mock.Anything, mock.AnythingOfType("domain.CreateVerificationParams")).
			Return(nil, verificationErr)

		// Act
		err := service.Register(ctx, name, email, password)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create verification")
	})

	t.Run("should return success when all operations succeed", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupIdentityService(t, config.AuthMethodPassword)
		ctx := context.Background()
		name := "John Doe"
		email := "john@example.com"
		password := "securePassword123"

		verification := createTestVerification(uuid.New(), domain.VerificationEmailFlow)

		deps.store.EXPECT().
			ExistsByEmail(ctx, email).
			Return(false, nil)

		deps.store.EXPECT().
			CreateUser(ctx, mock.AnythingOfType("sqlc.CreateUserParams")).
			Return(nil)

		deps.verificationService.EXPECT().
			CreateVerification(mock.Anything, mock.AnythingOfType("domain.CreateVerificationParams")).
			Return(verification, nil)

		deps.emailService.EXPECT().
			SendWelcomeVerificationEmailAsync(ctx, mock.AnythingOfType("*domain.User"), verification).
			Return()

		// Act
		err := service.Register(ctx, name, email, password)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should return success when auth method is both", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupIdentityService(t, config.AuthMethodBoth)
		ctx := context.Background()
		name := "Jane Doe"
		email := "jane@example.com"
		password := "anotherSecurePassword123"

		verification := createTestVerification(uuid.New(), domain.VerificationEmailFlow)

		deps.store.EXPECT().
			ExistsByEmail(ctx, email).
			Return(false, nil)

		deps.store.EXPECT().
			CreateUser(ctx, mock.AnythingOfType("sqlc.CreateUserParams")).
			Return(nil)

		deps.verificationService.EXPECT().
			CreateVerification(mock.Anything, mock.AnythingOfType("domain.CreateVerificationParams")).
			Return(verification, nil)

		deps.emailService.EXPECT().
			SendWelcomeVerificationEmailAsync(ctx, mock.AnythingOfType("*domain.User"), verification).
			Return()

		// Act
		err := service.Register(ctx, name, email, password)

		// Assert
		require.NoError(t, err)
	})
}

func TestRegisterMagicLink(t *testing.T) {
	t.Parallel()

	t.Run("should return ErrMagicLinkNotEnabled when magic link auth is disabled", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, _ := setupIdentityService(t, config.AuthMethodPassword)
		ctx := context.Background()
		name := "John Doe"
		email := "john@example.com"

		// Act
		err := service.RegisterMagicLink(ctx, name, email)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrMagicLinkNotEnabled)
	})

	t.Run("should return error when ExistsByEmail fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupIdentityService(t, config.AuthMethodMagicLink)
		ctx := context.Background()
		name := "John Doe"
		email := "john@example.com"
		dbErr := errors.New("database connection failed")

		deps.store.EXPECT().
			ExistsByEmail(ctx, email).
			Return(false, dbErr)

		// Act
		err := service.RegisterMagicLink(ctx, name, email)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "check email")
	})

	t.Run("should return ErrEmailAlreadyExists when email already exists", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupIdentityService(t, config.AuthMethodMagicLink)
		ctx := context.Background()
		name := "John Doe"
		email := "john@example.com"

		deps.store.EXPECT().
			ExistsByEmail(ctx, email).
			Return(true, nil)

		// Act
		err := service.RegisterMagicLink(ctx, name, email)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
	})

	t.Run("should return error when CreateUser fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupIdentityService(t, config.AuthMethodMagicLink)
		ctx := context.Background()
		name := "John Doe"
		email := "john@example.com"
		dbErr := errors.New("failed to create user")

		deps.store.EXPECT().
			ExistsByEmail(ctx, email).
			Return(false, nil)

		deps.store.EXPECT().
			CreateUser(ctx, mock.AnythingOfType("sqlc.CreateUserParams")).
			Return(dbErr)

		// Act
		err := service.RegisterMagicLink(ctx, name, email)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create user")
	})

	t.Run("should return error when CreateVerification fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupIdentityService(t, config.AuthMethodMagicLink)
		ctx := context.Background()
		name := "John Doe"
		email := "john@example.com"
		verificationErr := errors.New("failed to create verification")

		deps.store.EXPECT().
			ExistsByEmail(ctx, email).
			Return(false, nil)

		deps.store.EXPECT().
			CreateUser(ctx, mock.AnythingOfType("sqlc.CreateUserParams")).
			Return(nil)

		deps.verificationService.EXPECT().
			CreateVerification(mock.Anything, mock.AnythingOfType("domain.CreateVerificationParams")).
			Return(nil, verificationErr)

		// Act
		err := service.RegisterMagicLink(ctx, name, email)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create verification")
	})

	t.Run("should return success when all operations succeed", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupIdentityService(t, config.AuthMethodMagicLink)
		ctx := context.Background()
		name := "John Doe"
		email := "john@example.com"

		verification := createTestVerification(uuid.New(), domain.MagicLinkLoginFlow)

		deps.store.EXPECT().
			ExistsByEmail(ctx, email).
			Return(false, nil)

		deps.store.EXPECT().
			CreateUser(ctx, mock.AnythingOfType("sqlc.CreateUserParams")).
			Return(nil)

		deps.verificationService.EXPECT().
			CreateVerification(mock.Anything, mock.AnythingOfType("domain.CreateVerificationParams")).
			Return(verification, nil)

		deps.emailService.EXPECT().
			SendMagicLinkEmailAsync(ctx, mock.AnythingOfType("*domain.User"), verification).
			Return()

		// Act
		err := service.RegisterMagicLink(ctx, name, email)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should return success when auth method is both", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupIdentityService(t, config.AuthMethodBoth)
		ctx := context.Background()
		name := "Jane Doe"
		email := "jane@example.com"

		verification := createTestVerification(uuid.New(), domain.MagicLinkLoginFlow)

		deps.store.EXPECT().
			ExistsByEmail(ctx, email).
			Return(false, nil)

		deps.store.EXPECT().
			CreateUser(ctx, mock.AnythingOfType("sqlc.CreateUserParams")).
			Return(nil)

		deps.verificationService.EXPECT().
			CreateVerification(mock.Anything, mock.AnythingOfType("domain.CreateVerificationParams")).
			Return(verification, nil)

		deps.emailService.EXPECT().
			SendMagicLinkEmailAsync(ctx, mock.AnythingOfType("*domain.User"), verification).
			Return()

		// Act
		err := service.RegisterMagicLink(ctx, name, email)

		// Assert
		require.NoError(t, err)
	})
}

func TestConfirmEmail(t *testing.T) {
	t.Parallel()

	t.Run("should return error when ExchangeToken fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupIdentityService(t, config.AuthMethodPassword)
		ctx := context.Background()
		token := "invalid-token"
		exchangeErr := errors.New("token expired")

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.VerificationEmailFlow).
			Return(nil, exchangeErr)

		// Act
		result, err := service.ConfirmEmail(ctx, token)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "exchange token")
	})

	t.Run("should return error when VerifyUserEmail fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupIdentityService(t, config.AuthMethodPassword)
		ctx := context.Background()
		token := "valid-token"
		userID := uuid.New()
		dbErr := errors.New("database error")

		verification := createTestVerification(userID, domain.VerificationEmailFlow)

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.VerificationEmailFlow).
			Return(verification, nil)

		deps.store.EXPECT().
			VerifyUserEmail(ctx, mock.AnythingOfType("sqlc.VerifyUserEmailParams")).
			Return(dbErr)

		// Act
		result, err := service.ConfirmEmail(ctx, token)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "db verify email")
	})

	t.Run("should return success with correct UserID when all operations succeed", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupIdentityService(t, config.AuthMethodPassword)
		ctx := context.Background()
		token := "valid-token"
		userID := uuid.New()

		verification := createTestVerification(userID, domain.VerificationEmailFlow)

		deps.verificationService.EXPECT().
			ExchangeToken(ctx, token, domain.VerificationEmailFlow).
			Return(verification, nil)

		deps.store.EXPECT().
			VerifyUserEmail(ctx, mock.AnythingOfType("sqlc.VerifyUserEmailParams")).
			Return(nil)

		// Act
		result, err := service.ConfirmEmail(ctx, token)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
	})
}
