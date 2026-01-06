package service

import (
	"context"
	"log/slog"

	"github.com/gbvillarinho/base-project/internal/domain"
	"github.com/gbvillarinho/base-project/internal/infra/notification"
)

type EmailService interface {
	SendWelcomeVerificationEmailAsync(ctx context.Context, user *domain.User, verification *domain.Verification)
	SendVerifyEmailAsync(ctx context.Context, user *domain.User, verification *domain.Verification)
	SendChangeEmailAsync(ctx context.Context, user *domain.User, verification *domain.Verification)
	SendResetPasswordEmailAsync(ctx context.Context, user *domain.User, verification *domain.Verification)
	SendMagicLinkEmailAsync(ctx context.Context, user *domain.User, verification *domain.Verification)
}

type emailService struct {
	emailNotification notification.EmailNotification
	urlService        URLService
	logger            *slog.Logger
}

func NewEmailService(
	emailNotification notification.EmailNotification,
	urlService URLService,
	logger *slog.Logger,
) EmailService {
	return &emailService{
		emailNotification: emailNotification,
		urlService:        urlService,
		logger:            logger.With(slog.String("service", "email")),
	}
}

func (e *emailService) SendWelcomeVerificationEmailAsync(ctx context.Context, user *domain.User, verification *domain.Verification) {
	log := e.logger.With(
		slog.String("method", "SendWelcomeVerificationEmailAsync"),
		slog.String("user_id", user.ID.String()),
		slog.String("verification_id", verification.ID.String()),
	)

	if verification.Flow != domain.VerificationEmailFlow {
		log.Error("invalid verification flow for welcome email")
		return
	}

	params := notification.SendWelcomeEmailParams{
		UserRegistration: user.CreatedAt,
		Name:             user.Name,
		VerificationLink: e.urlService.GenerateVerificationURL(ctx, verification.Token, verification.Flow),
		Email:            user.Email,
	}

	go func() {
		ctx := context.Background()

		if err := e.emailNotification.SendWelcomeEmail(ctx, params); err != nil {
			log.Error("error to send welcome email", slog.String("error", err.Error()))
		}
	}()
}

func (e *emailService) SendVerifyEmailAsync(ctx context.Context, user *domain.User, verification *domain.Verification) {
	log := e.logger.With(
		slog.String("method", "SendVerifyEmailAsync"),
		slog.String("user_id", user.ID.String()),
		slog.String("verification_id", verification.ID.String()),
	)

	if verification.Flow != domain.VerificationEmailFlow {
		log.Error("invalid verification flow for verify email")
		return
	}

	params := notification.SendVerifyEmailParams{
		UserName:         user.Name,
		VerificationLink: e.urlService.GenerateVerificationURL(ctx, verification.Token, verification.Flow),
		UserEmail:        user.Email,
	}

	go func() {
		ctx := context.Background()

		if err := e.emailNotification.SendVerifyEmail(ctx, params); err != nil {
			log.Error("error to send verify email", slog.String("error", err.Error()))
		}
	}()
}

func (e *emailService) SendChangeEmailAsync(ctx context.Context, user *domain.User, verification *domain.Verification) {
	log := e.logger.With(
		slog.String("method", "SendChangeEmailAsync"),
		slog.String("user_id", user.ID.String()),
		slog.String("verification_id", verification.ID.String()),
	)

	if verification.Flow != domain.ChangeEmailFlow {
		log.Error("invalid verification flow for change email")
		return
	}

	params := notification.SendChangeEmailParams{
		UserName:         user.Name,
		NewEmail:         *verification.Payload,
		ConfirmationLink: e.urlService.GenerateVerificationURL(ctx, verification.Token, verification.Flow),
		UserEmail:        user.Email,
	}

	go func() {
		ctx := context.Background()

		if err := e.emailNotification.SendChangeEmailNotification(ctx, params); err != nil {
			log.Error("error to send verify email", slog.String("error", err.Error()))
		}
	}()
}

func (e *emailService) SendResetPasswordEmailAsync(ctx context.Context, user *domain.User, verification *domain.Verification) {
	log := e.logger.With(
		slog.String("method", "SendResetPasswordEmailAsync"),
		slog.String("user_id", user.ID.String()),
		slog.String("verification_id", verification.ID.String()),
	)

	if verification.Flow != domain.ResetPasswordFlow {
		log.Error("invalid verification flow for reset password")
		return
	}

	params := notification.SendResetPasswordEmailParams{
		UserName:  user.Name,
		ResetLink: e.urlService.GenerateVerificationURL(ctx, verification.Token, verification.Flow),
		UserEmail: user.Email,
	}

	go func() {
		ctx := context.Background()

		if err := e.emailNotification.SendResetPasswordEmail(ctx, params); err != nil {
			log.Error("error to send reset password email", slog.String("error", err.Error()))
		}
	}()
}

func (e *emailService) SendMagicLinkEmailAsync(ctx context.Context, user *domain.User, verification *domain.Verification) {
	log := e.logger.With(
		slog.String("method", "SendMagicLinkEmailAsync"),
		slog.String("user_id", user.ID.String()),
		slog.String("verification_id", verification.ID.String()),
	)

	if verification.Flow != domain.MagicLinkLoginFlow {
		log.Error("invalid verification flow for magic link email")
		return
	}

	params := notification.SendMagicLinkEmailParams{
		UserName:  user.Name,
		LoginLink: e.urlService.GenerateVerificationURL(ctx, verification.Token, verification.Flow),
		UserEmail: user.Email,
	}

	go func() {
		ctx := context.Background()

		if err := e.emailNotification.SendMagicLinkEmail(ctx, params); err != nil {
			log.Error("error to send magic link email", slog.String("error", err.Error()))
		}
	}()
}
