package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gbvillarinho/base-project/config"
	"github.com/gbvillarinho/base-project/internal/domain"
	"github.com/gbvillarinho/base-project/internal/infra/database/sqlc"
	"github.com/google/uuid"
)

type SessionService interface {
	CreateSession(ctx context.Context, userID uuid.UUID, ipAddress, deviceName, userAgent string) (*domain.Session, error)
	GetSessionByToken(ctx context.Context, token string) (*domain.Session, error)
	DeleteSessionByID(ctx context.Context, userID, sessionID uuid.UUID) error
	DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID, currentSession *uuid.UUID) error
	GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error)
}

type sessionService struct {
	store         sqlc.Store
	sessionConfig config.Session
}

func NewSessionService(
	store sqlc.Store,
	config *config.Config) SessionService {
	return &sessionService{
		store:         store,
		sessionConfig: config.Session,
	}
}

func (s *sessionService) CreateSession(ctx context.Context, userID uuid.UUID, ipAddress, deviceName, userAgent string) (*domain.Session, error) {
	expiresAt := time.Now().UTC().Add(s.sessionConfig.Duration)

	session, err := domain.NewSession(userID, ipAddress, userAgent, deviceName, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("create new session: %w", err)
	}

	if err := s.store.CreateSession(ctx, toCreateSessionParams(session)); err != nil {
		return nil, fmt.Errorf("persist new session: %w", err)
	}

	return session, nil
}

func (s *sessionService) GetSessionByToken(ctx context.Context, token string) (*domain.Session, error) {
	sessionFromDB, err := s.store.FindSessionByToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}

		return nil, fmt.Errorf("find session by token: %w", err)
	}

	session := toDomainSession(sessionFromDB)

	if session.IsExpired() {
		return nil, domain.ErrSessionExpired
	}

	return session, nil
}

func (s *sessionService) DeleteSessionByID(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	sessionFromDB, err := s.store.FindSessionByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrSessionNotFound
		}

		return fmt.Errorf("find session by id: %w", err)
	}
	session := toDomainSession(sessionFromDB)

	if session.UserID != userID {
		return domain.ErrSessionNotBelong
	}

	if session.IsExpired() {
		return nil
	}

	if err := s.store.DeleteSessionByID(ctx, sessionID); err != nil {
		return fmt.Errorf("delete session by id: %w", err)
	}

	return nil
}

func (s *sessionService) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID, currentSession *uuid.UUID) error {
	if currentSession == nil {
		if err := s.store.DeleteSessionsByUserID(ctx, userID); err != nil {
			return fmt.Errorf("delete all sessions by user id: %w", err)
		}

		return nil
	}

	params := sqlc.DeleteSessionsByUserExceptIDParams{
		UserID: userID,
		ID:     *currentSession,
	}

	if err := s.store.DeleteSessionsByUserExceptID(ctx, params); err != nil {
		return fmt.Errorf("delete all sessions by user id: %w", err)
	}

	return nil
}

func (s *sessionService) GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	params := sqlc.FindSessionsByUserIDParams{
		UserID:    userID,
		ExpiresAt: time.Now().UTC(),
	}

	sessionsFromDB, err := s.store.FindSessionsByUserID(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []domain.Session{}, nil
		}

		return nil, fmt.Errorf("get user sessions: %w", err)
	}

	sessions := make([]domain.Session, 0, len(sessionsFromDB))
	for _, session := range sessionsFromDB {
		sessions = append(sessions, *toDomainSession(session))
	}

	return sessions, nil
}
