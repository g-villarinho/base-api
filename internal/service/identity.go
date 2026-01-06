package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gbvillarinho/base-project/config"
	"github.com/gbvillarinho/base-project/internal/domain"
	"github.com/gbvillarinho/base-project/internal/infra/database/sqlc"
)

type IdentityService interface {
	Register(ctx context.Context, name, email, password string) error
	RegisterMagicLink(ctx context.Context, name, email string) error
	ConfirmEmail(ctx context.Context, token string) (*domain.ConfirmEmailResult, error)
}

type IdentityServiceImpl struct {
	store               sqlc.Store
	verificationService VerificationService
	emailService        EmailService
	config              *config.Config
}

func NewIdentityService(
	store sqlc.Store,
	verificationService VerificationService,
	emailService EmailService,
	config *config.Config) IdentityService {
	return &IdentityServiceImpl{
		store:               store,
		verificationService: verificationService,
		emailService:        emailService,
		config:              config,
	}
}

func (s *IdentityServiceImpl) Register(ctx context.Context, name, email, password string) error {
	if !s.config.IsPasswordAuthEnabled() {
		return domain.ErrPasswordAuthNotEnabled
	}

	emailAlreadyExists, err := s.store.ExistsByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("email already exists: %w", err)
	}

	if emailAlreadyExists {
		return domain.ErrEmailAlreadyExists
	}

	user := domain.NewUser(name, email, password)
	if err := s.store.CreateUser(ctx, toCreateUserParams(user)); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	verification, err := s.verificationService.CreateVerification(ctx, domain.CreateVerificationParams{
		UserID: user.ID,
		Flow:   domain.VerificationEmailFlow,
	})
	if err != nil {
		return fmt.Errorf("create verification: %w", err)
	}

	s.emailService.SendWelcomeVerificationEmailAsync(ctx, user, verification)

	return nil
}

func (s *IdentityServiceImpl) RegisterMagicLink(ctx context.Context, name, email string) error {
	if !s.config.IsMagicLinkAuthEnabled() {
		return domain.ErrMagicLinkNotEnabled
	}

	emailAlreadyExists, err := s.store.ExistsByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("check email: %w", err)
	}

	if emailAlreadyExists {
		return domain.ErrEmailAlreadyExists
	}

	user := domain.NewMagicLinkUser(name, email)
	if err := s.store.CreateUser(ctx, toCreateUserParams(user)); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	verification, err := s.verificationService.CreateVerification(ctx, domain.CreateVerificationParams{
		UserID: user.ID,
		Flow:   domain.MagicLinkLoginFlow,
	})
	if err != nil {
		return fmt.Errorf("create verification: %w", err)
	}

	s.emailService.SendMagicLinkEmailAsync(ctx, user, verification)

	return nil
}

func (s *IdentityServiceImpl) ConfirmEmail(ctx context.Context, token string) (*domain.ConfirmEmailResult, error) {
	verification, err := s.verificationService.ExchangeToken(ctx, token, domain.VerificationEmailFlow)
	if err != nil {
		return nil, fmt.Errorf("exchange token: %w", err)
	}

	now := time.Now().UTC()
	params := sqlc.VerifyUserEmailParams{
		UpdatedAt:        &now,
		EmailConfirmedAt: &now,
		ID:               verification.UserID,
	}

	if err := s.store.VerifyUserEmail(ctx, params); err != nil {
		return nil, fmt.Errorf("db verify email: %w", err)
	}

	result := &domain.ConfirmEmailResult{
		UserID: verification.UserID,
	}

	return result, nil
}
