package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gbvillarinho/base-project/internal/domain"
	"github.com/gbvillarinho/base-project/internal/infra/database/sqlc"
	"github.com/gbvillarinho/base-project/pkg/hash"
	"github.com/google/uuid"
)

type SecurityService interface {
	UpdatePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
	StartPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) (*domain.ResetPasswordResult, error)
	StartChangeEmail(ctx context.Context, userID uuid.UUID, newEmail string) error
	ChangeEmail(ctx context.Context, token string) error
}

type SecurityServiceImpl struct {
	store               sqlc.Store
	verificationService VerificationService
	emailService        EmailService
}

func NewSecurityService(
	store sqlc.Store,
	verificationService VerificationService,
	emailService EmailService) SecurityService {
	return &SecurityServiceImpl{
		store:               store,
		verificationService: verificationService,
		emailService:        emailService,
	}
}

func (s *SecurityServiceImpl) UpdatePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	userFromDB, err := s.store.FindUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("could not find user: %w", err)
	}

	user := toDomainUser(userFromDB)

	if !user.HasPassword() || hash.VerifyPassword(currentPassword, *user.PasswordHash) != nil {
		return domain.ErrPasswordMismatch
	}

	newPasswordHash, err := hash.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("could not hash new password: %w", err)
	}

	now := time.Now().UTC()
	params := sqlc.UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: newPasswordHash,
		UpdatedAt:    &now,
	}

	if err := s.store.UpdateUserPassword(ctx, params); err != nil {
		return fmt.Errorf("could not update user password: %w", err)
	}

	return nil
}

func (s *SecurityServiceImpl) StartPasswordReset(ctx context.Context, email string) error {
	userFromDB, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // Do not reveal if email exists (prevents enumeration)
		}

		return fmt.Errorf("find user: %w", err)
	}

	user := toDomainUser(userFromDB)
	verification, err := s.verificationService.CreateVerification(ctx, domain.CreateVerificationParams{
		UserID: user.ID,
		Flow:   domain.ResetPasswordFlow,
	})
	if err != nil {
		return fmt.Errorf("create verification: %w", err)
	}

	s.emailService.SendResetPasswordEmailAsync(ctx, user, verification)
	return nil
}

func (s *SecurityServiceImpl) ResetPassword(ctx context.Context, token, newPassword string) (*domain.ResetPasswordResult, error) {
	verification, err := s.verificationService.ExchangeToken(ctx, token, domain.ResetPasswordFlow)
	if err != nil {
		return nil, fmt.Errorf("consume token: %w", err)
	}

	newPasswordHash, err := hash.HashPassword(newPassword)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	params := sqlc.UpdateUserPasswordParams{
		PasswordHash: newPasswordHash,
		UpdatedAt:    &now,
		ID:           verification.UserID,
	}

	if err := s.store.UpdateUserPassword(ctx, params); err != nil {
		return nil, fmt.Errorf("update db password: %w", err)
	}

	return &domain.ResetPasswordResult{
		UserID: verification.UserID,
	}, nil
}

func (s *SecurityServiceImpl) StartChangeEmail(ctx context.Context, userID uuid.UUID, newEmail string) error {
	userFromDB, err := s.store.FindUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user by id: %w", err)
	}

	if userFromDB.Email == newEmail {
		return domain.ErrEmailIsTheSame
	}

	emailAlreadyExists, err := s.store.ExistsByEmail(ctx, newEmail)
	if err != nil {
		return fmt.Errorf("check email existence: %w", err)
	}

	if emailAlreadyExists {
		return domain.ErrEmailAlreadyExists
	}

	user := toDomainUser(userFromDB)
	verification, err := s.verificationService.CreateVerification(ctx, domain.CreateVerificationParams{
		UserID:  user.ID,
		Flow:    domain.ChangeEmailFlow,
		Payload: newEmail,
	})
	if err != nil {
		return fmt.Errorf("create verification: %w", err)
	}

	s.emailService.SendChangeEmailAsync(ctx, user, verification)
	return nil
}

func (s *SecurityServiceImpl) ChangeEmail(ctx context.Context, token string) error {
	verification, err := s.verificationService.ExchangeToken(ctx, token, domain.ChangeEmailFlow)
	if err != nil {
		return fmt.Errorf("consume token: %w", err)
	}

	if verification.Payload == nil {
		return domain.ErrInvalidVerificationPayload
	}

	now := time.Now().UTC()
	params := sqlc.UpdateUserEmailParams{
		Email:     *verification.Payload,
		UpdatedAt: &now,
		ID:        verification.UserID,
	}

	if err := s.store.UpdateUserEmail(ctx, params); err != nil {
		return fmt.Errorf("update db email: %w", err)
	}

	return nil
}
