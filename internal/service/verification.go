package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gbvillarinho/base-project/internal/domain"
	"github.com/gbvillarinho/base-project/internal/infra/database/sqlc"
)

const (
	VerifyEmailExpirationMinute = 10 * time.Minute
	MinTimeRemaining            = 5 * time.Minute
)

type VerificationService interface {
	CreateVerification(ctx context.Context, params domain.CreateVerificationParams) (*domain.Verification, error)
	ExchangeToken(ctx context.Context, token string, expectedFlow domain.VerificationFlow) (*domain.Verification, error)
}

type verificationService struct {
	store  sqlc.Store
	logger *slog.Logger
}

func NewVerificationService(
	store sqlc.Store,
) VerificationService {
	return &verificationService{
		store: store,
	}
}

func (s *verificationService) CreateVerification(ctx context.Context, params domain.CreateVerificationParams) (*domain.Verification, error) {
	existingParams := sqlc.FindValidVerificationByUserIDAndFlowParams{
		UserID:    params.UserID,
		Flow:      string(params.Flow),
		ExpiresAt: time.Now().UTC(),
	}

	existingVerification, err := s.store.FindValidVerificationByUserIDAndFlow(ctx, existingParams)
	if err == nil {
		timeRemaining := time.Until(existingVerification.ExpiresAt)

		if timeRemaining > MinTimeRemaining {
			return toDomainVerification(existingVerification), nil
		}

	}

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("find valid verification: %w", err)
	}

	expiresAt := time.Now().UTC().Add(VerifyEmailExpirationMinute)
	verification, err := domain.NewVerification(params.UserID, params.Flow, expiresAt, params.Payload)
	if err != nil {
		return nil, fmt.Errorf("create verification entity: %w", err)
	}

	if err := s.store.CreateVerification(ctx, toCreateVerificationParams(verification)); err != nil {
		return nil, fmt.Errorf("create verification: %w", err)
	}

	return verification, nil
}

func (s *verificationService) ExchangeToken(ctx context.Context, token string, expectedFlow domain.VerificationFlow) (*domain.Verification, error) {
	verificationFromToken, err := s.store.FindVerificationByToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidVerification
		}

		return nil, fmt.Errorf("find verification by token: %w", err)
	}

	verification := toDomainVerification(verificationFromToken)

	if verification.IsExpired() {
		return nil, domain.ErrInvalidVerification
	}

	if verification.Flow != expectedFlow {
		return nil, domain.ErrInvalidVerification
	}

	if err := s.store.DeleteVerification(ctx, verification.ID); err != nil {
		return nil, fmt.Errorf("delete verification: %w", err)
	}

	return verification, nil
}
