package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/g-villarinho/voxel-api/config"
	"github.com/g-villarinho/voxel-api/internal/domain"
	"github.com/g-villarinho/voxel-api/internal/infra/database/sqlc"
	"github.com/g-villarinho/voxel-api/pkg/hash"
)

type LoginService interface {
	Authenticate(ctx context.Context, email, password string) (*domain.User, error)
	RequestMagicLink(ctx context.Context, email string) error
	VerifyMagicLink(ctx context.Context, token string) (*domain.User, error)
}

type LoginServiceImpl struct {
	store               sqlc.Store
	verificationService VerificationService
	emailService        EmailService
	config              *config.Config
}

func NewLoginService(
	store sqlc.Store,
	verificationService VerificationService,
	emailService EmailService,
	config *config.Config) LoginService {
	return &LoginServiceImpl{
		store:               store,
		verificationService: verificationService,
		emailService:        emailService,
		config:              config,
	}
}

func (s *LoginServiceImpl) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	if !s.config.IsPasswordAuthEnabled() {
		return nil, domain.ErrPasswordAuthNotEnabled
	}

	userFromEmail, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidCredentials
		}

		return nil, fmt.Errorf("find user by email: %w", err)
	}

	user := toDomainUser(userFromEmail)
	if !user.HasPassword() || hash.VerifyPassword(password, *user.PasswordHash) != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if user.IsBlocked() {
		return nil, domain.ErrUserBlocked
	}

	if !user.IsEmailVerified() {
		return nil, domain.ErrEmailNotVerified
	}

	return user, nil
}

func (s *LoginServiceImpl) RequestMagicLink(ctx context.Context, email string) error {
	if !s.config.IsMagicLinkAuthEnabled() {
		return domain.ErrMagicLinkNotEnabled
	}

	userFromEmail, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // Do not reveal that the email does not exist
		}

		return fmt.Errorf("find user: %w", err)
	}

	user := toDomainUser(userFromEmail)

	if user.IsBlocked() {
		return nil // Do not reveal that the user is blocked
	}

	verification, err := s.verificationService.CreateVerification(ctx, domain.CreateVerificationParams{
		UserID: user.ID,
		Flow:   domain.MagicLinkLoginFlow,
	})
	if err != nil {
		return fmt.Errorf("create magic link: %w", err)
	}

	s.emailService.SendMagicLinkEmailAsync(ctx, user, verification)
	return nil
}

func (s *LoginServiceImpl) VerifyMagicLink(ctx context.Context, token string) (*domain.User, error) {
	verification, err := s.verificationService.ExchangeToken(ctx, token, domain.MagicLinkLoginFlow)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	userFromDB, err := s.store.FindUserByID(ctx, verification.UserID)
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	user := toDomainUser(userFromDB)

	if user.IsBlocked() {
		return nil, domain.ErrUserBlocked
	}

	if !user.IsEmailVerified() {
		now := time.Now().UTC()
		params := sqlc.VerifyUserEmailParams{
			UpdatedAt:        &now,
			EmailConfirmedAt: &now,
			ID:               user.ID,
		}

		if err := s.store.VerifyUserEmail(ctx, params); err != nil {
			return nil, fmt.Errorf("verify user email: %w", err)
		}
	}

	return user, nil
}
