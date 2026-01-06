package notification_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/g-villarinho/voxel-api/internal/infra/notification"
	"github.com/g-villarinho/voxel-api/internal/mocks"
)

func setupEmailNotification(t *testing.T) (notification.EmailNotification, *mocks.EmailClientMock) {
	t.Helper()

	emailClientMock := mocks.NewEmailClientMock(t)
	emailNotification := notification.NewEmailNotification(emailClientMock)

	return emailNotification, emailClientMock
}

func TestSendWelcomeEmail(t *testing.T) {
	t.Run("should send welcome email successfully", func(t *testing.T) {
		// Arrange
		emailNotification, emailClientMock := setupEmailNotification(t)
		ctx := context.Background()

		params := notification.SendWelcomeEmailParams{
			UserRegistration: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			Name:             "John Doe",
			VerificationLink: "https://example.com/verify?token=abc123",
			Email:            "john@example.com",
		}

		emailClientMock.EXPECT().
			SendEmail(ctx, params.Email, "Welcome to voxel! Please, verify your email", mock.AnythingOfType("string")).
			Return(nil)

		// Act
		err := emailNotification.SendWelcomeEmail(ctx, params)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should return error when email client fails", func(t *testing.T) {
		// Arrange
		emailNotification, emailClientMock := setupEmailNotification(t)
		ctx := context.Background()

		params := notification.SendWelcomeEmailParams{
			UserRegistration: time.Now(),
			Name:             "John Doe",
			VerificationLink: "https://example.com/verify?token=abc123",
			Email:            "john@example.com",
		}

		emailClientMock.EXPECT().
			SendEmail(ctx, params.Email, "Welcome to voxel! Please, verify your email", mock.AnythingOfType("string")).
			Return(errors.New("email service unavailable"))

		// Act
		err := emailNotification.SendWelcomeEmail(ctx, params)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "send email")
		assert.Contains(t, err.Error(), "email service unavailable")
	})

	t.Run("should include user data in email body", func(t *testing.T) {
		// Arrange
		emailNotification, emailClientMock := setupEmailNotification(t)
		ctx := context.Background()

		params := notification.SendWelcomeEmailParams{
			UserRegistration: time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
			Name:             "Jane Smith",
			VerificationLink: "https://example.com/verify?token=xyz789",
			Email:            "jane@example.com",
		}

		var capturedBody string
		emailClientMock.EXPECT().
			SendEmail(ctx, params.Email, "Welcome to voxel! Please, verify your email", mock.AnythingOfType("string")).
			Run(func(ctx context.Context, to, subject, body string) {
				capturedBody = body
			}).
			Return(nil)

		// Act
		err := emailNotification.SendWelcomeEmail(ctx, params)

		// Assert
		require.NoError(t, err)
		assert.Contains(t, capturedBody, params.Name)
		assert.Contains(t, capturedBody, params.VerificationLink)
		assert.Contains(t, capturedBody, params.Email)
	})
}

func TestSendVerifyEmail(t *testing.T) {
	t.Run("should send verify email successfully", func(t *testing.T) {
		// Arrange
		emailNotification, emailClientMock := setupEmailNotification(t)
		ctx := context.Background()

		params := notification.SendVerifyEmailParams{
			UserName:         "John Doe",
			VerificationLink: "https://example.com/verify?token=abc123",
			UserEmail:        "john@example.com",
		}

		emailClientMock.EXPECT().
			SendEmail(ctx, params.UserEmail, "Verify your email address", mock.AnythingOfType("string")).
			Return(nil)

		// Act
		err := emailNotification.SendVerifyEmail(ctx, params)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should return error when email client fails", func(t *testing.T) {
		// Arrange
		emailNotification, emailClientMock := setupEmailNotification(t)
		ctx := context.Background()

		params := notification.SendVerifyEmailParams{
			UserName:         "John Doe",
			VerificationLink: "https://example.com/verify?token=abc123",
			UserEmail:        "john@example.com",
		}

		emailClientMock.EXPECT().
			SendEmail(ctx, params.UserEmail, "Verify your email address", mock.AnythingOfType("string")).
			Return(errors.New("connection timeout"))

		// Act
		err := emailNotification.SendVerifyEmail(ctx, params)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connection timeout")
	})

	t.Run("should include verification link in email body", func(t *testing.T) {
		// Arrange
		emailNotification, emailClientMock := setupEmailNotification(t)
		ctx := context.Background()

		params := notification.SendVerifyEmailParams{
			UserName:         "Jane Smith",
			VerificationLink: "https://example.com/verify?token=unique-token",
			UserEmail:        "jane@example.com",
		}

		var capturedBody string
		emailClientMock.EXPECT().
			SendEmail(ctx, params.UserEmail, "Verify your email address", mock.AnythingOfType("string")).
			Run(func(ctx context.Context, to, subject, body string) {
				capturedBody = body
			}).
			Return(nil)

		// Act
		err := emailNotification.SendVerifyEmail(ctx, params)

		// Assert
		require.NoError(t, err)
		assert.Contains(t, capturedBody, params.UserName)
		assert.Contains(t, capturedBody, params.VerificationLink)
		assert.Contains(t, capturedBody, params.UserEmail)
	})
}

func TestSendResetPasswordEmail(t *testing.T) {
	t.Run("should send reset password email successfully", func(t *testing.T) {
		// Arrange
		emailNotification, emailClientMock := setupEmailNotification(t)
		ctx := context.Background()

		params := notification.SendResetPasswordEmailParams{
			UserName:  "John Doe",
			ResetLink: "https://example.com/reset?token=abc123",
			UserEmail: "john@example.com",
		}

		emailClientMock.EXPECT().
			SendEmail(ctx, params.UserEmail, "Password Reset Request", mock.AnythingOfType("string")).
			Return(nil)

		// Act
		err := emailNotification.SendResetPasswordEmail(ctx, params)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should return error when email client fails", func(t *testing.T) {
		// Arrange
		emailNotification, emailClientMock := setupEmailNotification(t)
		ctx := context.Background()

		params := notification.SendResetPasswordEmailParams{
			UserName:  "John Doe",
			ResetLink: "https://example.com/reset?token=abc123",
			UserEmail: "john@example.com",
		}

		emailClientMock.EXPECT().
			SendEmail(ctx, params.UserEmail, "Password Reset Request", mock.AnythingOfType("string")).
			Return(errors.New("rate limit exceeded"))

		// Act
		err := emailNotification.SendResetPasswordEmail(ctx, params)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit exceeded")
	})

	t.Run("should include reset link in email body", func(t *testing.T) {
		// Arrange
		emailNotification, emailClientMock := setupEmailNotification(t)
		ctx := context.Background()

		params := notification.SendResetPasswordEmailParams{
			UserName:  "Jane Smith",
			ResetLink: "https://example.com/reset?token=secure-reset-token",
			UserEmail: "jane@example.com",
		}

		var capturedBody string
		emailClientMock.EXPECT().
			SendEmail(ctx, params.UserEmail, "Password Reset Request", mock.AnythingOfType("string")).
			Run(func(ctx context.Context, to, subject, body string) {
				capturedBody = body
			}).
			Return(nil)

		// Act
		err := emailNotification.SendResetPasswordEmail(ctx, params)

		// Assert
		require.NoError(t, err)
		assert.Contains(t, capturedBody, params.UserName)
		assert.Contains(t, capturedBody, params.ResetLink)
		assert.Contains(t, capturedBody, params.UserEmail)
	})
}

func TestSendChangeEmailNotification(t *testing.T) {
	t.Run("should send change email notification successfully", func(t *testing.T) {
		// Arrange
		emailNotification, emailClientMock := setupEmailNotification(t)
		ctx := context.Background()

		params := notification.SendChangeEmailParams{
			UserName:         "John Doe",
			NewEmail:         "john.new@example.com",
			ConfirmationLink: "https://example.com/confirm-email?token=abc123",
			UserEmail:        "john@example.com",
		}

		emailClientMock.EXPECT().
			SendEmail(ctx, params.UserEmail, "Email Change Confirmation", mock.AnythingOfType("string")).
			Return(nil)

		// Act
		err := emailNotification.SendChangeEmailNotification(ctx, params)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should return error when email client fails", func(t *testing.T) {
		// Arrange
		emailNotification, emailClientMock := setupEmailNotification(t)
		ctx := context.Background()

		params := notification.SendChangeEmailParams{
			UserName:         "John Doe",
			NewEmail:         "john.new@example.com",
			ConfirmationLink: "https://example.com/confirm-email?token=abc123",
			UserEmail:        "john@example.com",
		}

		emailClientMock.EXPECT().
			SendEmail(ctx, params.UserEmail, "Email Change Confirmation", mock.AnythingOfType("string")).
			Return(errors.New("invalid recipient"))

		// Act
		err := emailNotification.SendChangeEmailNotification(ctx, params)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid recipient")
	})

	t.Run("should include new email and confirmation link in email body", func(t *testing.T) {
		// Arrange
		emailNotification, emailClientMock := setupEmailNotification(t)
		ctx := context.Background()

		params := notification.SendChangeEmailParams{
			UserName:         "Jane Smith",
			NewEmail:         "jane.updated@example.com",
			ConfirmationLink: "https://example.com/confirm-email?token=change-token",
			UserEmail:        "jane@example.com",
		}

		var capturedBody string
		emailClientMock.EXPECT().
			SendEmail(ctx, params.UserEmail, "Email Change Confirmation", mock.AnythingOfType("string")).
			Run(func(ctx context.Context, to, subject, body string) {
				capturedBody = body
			}).
			Return(nil)

		// Act
		err := emailNotification.SendChangeEmailNotification(ctx, params)

		// Assert
		require.NoError(t, err)
		assert.Contains(t, capturedBody, params.UserName)
		assert.Contains(t, capturedBody, params.NewEmail)
		assert.Contains(t, capturedBody, params.ConfirmationLink)
		assert.Contains(t, capturedBody, params.UserEmail)
	})
}

func TestSendMagicLinkEmail(t *testing.T) {
	t.Run("should send magic link email successfully", func(t *testing.T) {
		// Arrange
		emailNotification, emailClientMock := setupEmailNotification(t)
		ctx := context.Background()

		params := notification.SendMagicLinkEmailParams{
			UserName:  "John Doe",
			LoginLink: "https://example.com/login?token=magic-abc123",
			UserEmail: "john@example.com",
		}

		emailClientMock.EXPECT().
			SendEmail(ctx, params.UserEmail, "Your login link", mock.AnythingOfType("string")).
			Return(nil)

		// Act
		err := emailNotification.SendMagicLinkEmail(ctx, params)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should return error when email client fails", func(t *testing.T) {
		// Arrange
		emailNotification, emailClientMock := setupEmailNotification(t)
		ctx := context.Background()

		params := notification.SendMagicLinkEmailParams{
			UserName:  "John Doe",
			LoginLink: "https://example.com/login?token=magic-abc123",
			UserEmail: "john@example.com",
		}

		emailClientMock.EXPECT().
			SendEmail(ctx, params.UserEmail, "Your login link", mock.AnythingOfType("string")).
			Return(errors.New("smtp connection failed"))

		// Act
		err := emailNotification.SendMagicLinkEmail(ctx, params)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "send email")
		assert.Contains(t, err.Error(), "smtp connection failed")
	})

	t.Run("should include login link in email body", func(t *testing.T) {
		// Arrange
		emailNotification, emailClientMock := setupEmailNotification(t)
		ctx := context.Background()

		params := notification.SendMagicLinkEmailParams{
			UserName:  "Jane Smith",
			LoginLink: "https://example.com/login?token=secure-magic-link",
			UserEmail: "jane@example.com",
		}

		var capturedBody string
		emailClientMock.EXPECT().
			SendEmail(ctx, params.UserEmail, "Your login link", mock.AnythingOfType("string")).
			Run(func(ctx context.Context, to, subject, body string) {
				capturedBody = body
			}).
			Return(nil)

		// Act
		err := emailNotification.SendMagicLinkEmail(ctx, params)

		// Assert
		require.NoError(t, err)
		assert.Contains(t, capturedBody, params.UserName)
		assert.Contains(t, capturedBody, params.LoginLink)
		assert.Contains(t, capturedBody, params.UserEmail)
	})
}
