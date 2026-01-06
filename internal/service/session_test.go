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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type sessionTestDeps struct {
	store  *mocks.StoreMock
	config *config.Config
}

func setupSessionService(t *testing.T) (SessionService, *sessionTestDeps) {
	t.Helper()

	deps := &sessionTestDeps{
		store: mocks.NewStoreMock(t),
		config: &config.Config{
			Session: config.Session{
				Duration: 24 * time.Hour,
			},
		},
	}

	service := NewSessionService(deps.store, deps.config)

	return service, deps
}

func createTestSQLCSession(id, userID uuid.UUID, token, ipAddress, userAgent, deviceName string, expiresAt time.Time) sqlc.Session {
	return sqlc.Session{
		ID:         id,
		UserID:     userID,
		Token:      token,
		IpAddress:  ipAddress,
		UserAgent:  userAgent,
		DeviceName: deviceName,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now().UTC(),
	}
}

func TestCreateSession(t *testing.T) {
	t.Parallel()

	t.Run("should create session successfully", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		userID := uuid.New()
		ipAddress := "192.168.1.1"
		deviceName := "Chrome on Windows"
		userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"

		deps.store.EXPECT().
			CreateSession(ctx, mock.AnythingOfType("sqlc.CreateSessionParams")).
			Return(nil)

		// Act
		session, err := service.CreateSession(ctx, userID, ipAddress, deviceName, userAgent)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, userID, session.UserID)
		assert.Equal(t, ipAddress, session.IPAddress)
		assert.Equal(t, deviceName, session.DeviceName)
		assert.Equal(t, userAgent, session.UserAgent)
		assert.NotEmpty(t, session.Token)
		assert.NotEqual(t, uuid.Nil, session.ID)
		assert.False(t, session.ExpiresAt.IsZero())
	})

	t.Run("should return error when store fails to create session", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		userID := uuid.New()
		ipAddress := "192.168.1.1"
		deviceName := "Chrome on Windows"
		userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
		dbErr := errors.New("database connection error")

		deps.store.EXPECT().
			CreateSession(ctx, mock.AnythingOfType("sqlc.CreateSessionParams")).
			Return(dbErr)

		// Act
		session, err := service.CreateSession(ctx, userID, ipAddress, deviceName, userAgent)

		// Assert
		require.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "persist new session")
	})
}

func TestGetSessionByToken(t *testing.T) {
	t.Parallel()

	t.Run("should return session when token is valid and not expired", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		sessionID := uuid.New()
		userID := uuid.New()
		token := "valid-session-token"
		expiresAt := time.Now().UTC().Add(24 * time.Hour)

		sqlcSession := createTestSQLCSession(sessionID, userID, token, "192.168.1.1", "Mozilla/5.0", "Chrome", expiresAt)

		deps.store.EXPECT().
			FindSessionByToken(ctx, token).
			Return(sqlcSession, nil)

		// Act
		session, err := service.GetSessionByToken(ctx, token)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, sessionID, session.ID)
		assert.Equal(t, userID, session.UserID)
		assert.Equal(t, token, session.Token)
	})

	t.Run("should return ErrSessionNotFound when session does not exist", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		token := "nonexistent-token"

		deps.store.EXPECT().
			FindSessionByToken(ctx, token).
			Return(sqlc.Session{}, sql.ErrNoRows)

		// Act
		session, err := service.GetSessionByToken(ctx, token)

		// Assert
		require.Error(t, err)
		assert.Nil(t, session)
		assert.ErrorIs(t, err, domain.ErrSessionNotFound)
	})

	t.Run("should return error when database fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		token := "some-token"
		dbErr := errors.New("database connection error")

		deps.store.EXPECT().
			FindSessionByToken(ctx, token).
			Return(sqlc.Session{}, dbErr)

		// Act
		session, err := service.GetSessionByToken(ctx, token)

		// Assert
		require.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "find session by token")
	})

	t.Run("should return ErrSessionExpired when session is expired", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		sessionID := uuid.New()
		userID := uuid.New()
		token := "expired-token"
		expiresAt := time.Now().UTC().Add(-1 * time.Hour) // expired 1 hour ago

		sqlcSession := createTestSQLCSession(sessionID, userID, token, "192.168.1.1", "Mozilla/5.0", "Chrome", expiresAt)

		deps.store.EXPECT().
			FindSessionByToken(ctx, token).
			Return(sqlcSession, nil)

		// Act
		session, err := service.GetSessionByToken(ctx, token)

		// Assert
		require.Error(t, err)
		assert.Nil(t, session)
		assert.ErrorIs(t, err, domain.ErrSessionExpired)
	})
}

func TestDeleteSessionByID(t *testing.T) {
	t.Parallel()

	t.Run("should delete session successfully when session belongs to user", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		sessionID := uuid.New()
		userID := uuid.New()
		expiresAt := time.Now().UTC().Add(24 * time.Hour)

		sqlcSession := createTestSQLCSession(sessionID, userID, "token", "192.168.1.1", "Mozilla/5.0", "Chrome", expiresAt)

		deps.store.EXPECT().
			FindSessionByID(ctx, sessionID).
			Return(sqlcSession, nil)

		deps.store.EXPECT().
			DeleteSessionByID(ctx, sessionID).
			Return(nil)

		// Act
		err := service.DeleteSessionByID(ctx, userID, sessionID)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should return ErrSessionNotFound when session does not exist", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		sessionID := uuid.New()
		userID := uuid.New()

		deps.store.EXPECT().
			FindSessionByID(ctx, sessionID).
			Return(sqlc.Session{}, sql.ErrNoRows)

		// Act
		err := service.DeleteSessionByID(ctx, userID, sessionID)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrSessionNotFound)
	})

	t.Run("should return error when database fails to find session", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		sessionID := uuid.New()
		userID := uuid.New()
		dbErr := errors.New("database connection error")

		deps.store.EXPECT().
			FindSessionByID(ctx, sessionID).
			Return(sqlc.Session{}, dbErr)

		// Act
		err := service.DeleteSessionByID(ctx, userID, sessionID)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "find session by id")
	})

	t.Run("should return ErrSessionNotBelong when session belongs to different user", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		sessionID := uuid.New()
		sessionOwnerID := uuid.New()
		differentUserID := uuid.New()
		expiresAt := time.Now().UTC().Add(24 * time.Hour)

		sqlcSession := createTestSQLCSession(sessionID, sessionOwnerID, "token", "192.168.1.1", "Mozilla/5.0", "Chrome", expiresAt)

		deps.store.EXPECT().
			FindSessionByID(ctx, sessionID).
			Return(sqlcSession, nil)

		// Act
		err := service.DeleteSessionByID(ctx, differentUserID, sessionID)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrSessionNotBelong)
	})

	t.Run("should return nil when session is expired", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		sessionID := uuid.New()
		userID := uuid.New()
		expiresAt := time.Now().UTC().Add(-1 * time.Hour) // expired 1 hour ago

		sqlcSession := createTestSQLCSession(sessionID, userID, "token", "192.168.1.1", "Mozilla/5.0", "Chrome", expiresAt)

		deps.store.EXPECT().
			FindSessionByID(ctx, sessionID).
			Return(sqlcSession, nil)

		// Act
		err := service.DeleteSessionByID(ctx, userID, sessionID)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should return error when database fails to delete session", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		sessionID := uuid.New()
		userID := uuid.New()
		expiresAt := time.Now().UTC().Add(24 * time.Hour)
		dbErr := errors.New("database connection error")

		sqlcSession := createTestSQLCSession(sessionID, userID, "token", "192.168.1.1", "Mozilla/5.0", "Chrome", expiresAt)

		deps.store.EXPECT().
			FindSessionByID(ctx, sessionID).
			Return(sqlcSession, nil)

		deps.store.EXPECT().
			DeleteSessionByID(ctx, sessionID).
			Return(dbErr)

		// Act
		err := service.DeleteSessionByID(ctx, userID, sessionID)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "delete session by id")
	})
}

func TestDeleteSessionsByUserID(t *testing.T) {
	t.Parallel()

	t.Run("should delete all sessions when currentSession is nil", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		userID := uuid.New()

		deps.store.EXPECT().
			DeleteSessionsByUserID(ctx, userID).
			Return(nil)

		// Act
		err := service.DeleteSessionsByUserID(ctx, userID, nil)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should return error when database fails to delete all sessions", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		userID := uuid.New()
		dbErr := errors.New("database connection error")

		deps.store.EXPECT().
			DeleteSessionsByUserID(ctx, userID).
			Return(dbErr)

		// Act
		err := service.DeleteSessionsByUserID(ctx, userID, nil)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "delete all sessions by user id")
	})

	t.Run("should delete all sessions except current when currentSession is provided", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		userID := uuid.New()
		currentSessionID := uuid.New()

		expectedParams := sqlc.DeleteSessionsByUserExceptIDParams{
			UserID: userID,
			ID:     currentSessionID,
		}

		deps.store.EXPECT().
			DeleteSessionsByUserExceptID(ctx, expectedParams).
			Return(nil)

		// Act
		err := service.DeleteSessionsByUserID(ctx, userID, &currentSessionID)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should return error when database fails to delete sessions except current", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		userID := uuid.New()
		currentSessionID := uuid.New()
		dbErr := errors.New("database connection error")

		expectedParams := sqlc.DeleteSessionsByUserExceptIDParams{
			UserID: userID,
			ID:     currentSessionID,
		}

		deps.store.EXPECT().
			DeleteSessionsByUserExceptID(ctx, expectedParams).
			Return(dbErr)

		// Act
		err := service.DeleteSessionsByUserID(ctx, userID, &currentSessionID)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "delete all sessions by user id")
	})
}

func TestGetSessionsByUserID(t *testing.T) {
	t.Parallel()

	t.Run("should return sessions successfully", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		userID := uuid.New()
		session1ID := uuid.New()
		session2ID := uuid.New()
		expiresAt := time.Now().UTC().Add(24 * time.Hour)

		sqlcSessions := []sqlc.Session{
			createTestSQLCSession(session1ID, userID, "token1", "192.168.1.1", "Mozilla/5.0", "Chrome", expiresAt),
			createTestSQLCSession(session2ID, userID, "token2", "192.168.1.2", "Safari", "Safari on Mac", expiresAt),
		}

		deps.store.EXPECT().
			FindSessionsByUserID(ctx, mock.AnythingOfType("sqlc.FindSessionsByUserIDParams")).
			Return(sqlcSessions, nil)

		// Act
		sessions, err := service.GetSessionsByUserID(ctx, userID)

		// Assert
		require.NoError(t, err)
		assert.Len(t, sessions, 2)
		assert.Equal(t, session1ID, sessions[0].ID)
		assert.Equal(t, session2ID, sessions[1].ID)
	})

	t.Run("should return empty slice when no sessions exist", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		userID := uuid.New()

		deps.store.EXPECT().
			FindSessionsByUserID(ctx, mock.AnythingOfType("sqlc.FindSessionsByUserIDParams")).
			Return([]sqlc.Session{}, sql.ErrNoRows)

		// Act
		sessions, err := service.GetSessionsByUserID(ctx, userID)

		// Assert
		require.NoError(t, err)
		assert.Empty(t, sessions)
	})

	t.Run("should return error when database fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		userID := uuid.New()
		dbErr := errors.New("database connection error")

		deps.store.EXPECT().
			FindSessionsByUserID(ctx, mock.AnythingOfType("sqlc.FindSessionsByUserIDParams")).
			Return(nil, dbErr)

		// Act
		sessions, err := service.GetSessionsByUserID(ctx, userID)

		// Assert
		require.Error(t, err)
		assert.Nil(t, sessions)
		assert.Contains(t, err.Error(), "get user sessions")
	})

	t.Run("should return empty slice when store returns empty result", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service, deps := setupSessionService(t)
		ctx := context.Background()
		userID := uuid.New()

		deps.store.EXPECT().
			FindSessionsByUserID(ctx, mock.AnythingOfType("sqlc.FindSessionsByUserIDParams")).
			Return([]sqlc.Session{}, nil)

		// Act
		sessions, err := service.GetSessionsByUserID(ctx, userID)

		// Assert
		require.NoError(t, err)
		assert.Empty(t, sessions)
	})
}
