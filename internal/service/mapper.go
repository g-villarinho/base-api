package service

import (
	"github.com/g-villarinho/voxel-api/internal/domain"
	"github.com/g-villarinho/voxel-api/internal/infra/database/sqlc"
)

func toDomainUser(dbUser sqlc.User) *domain.User {
	var passwordHash *string
	if dbUser.PasswordHash != "" {
		passwordHash = &dbUser.PasswordHash
	}

	return &domain.User{
		ID:               dbUser.ID,
		Name:             dbUser.Name,
		Email:            dbUser.Email,
		Status:           domain.UserStatus(dbUser.Status),
		PasswordHash:     passwordHash,
		CreatedAt:        dbUser.CreatedAt,
		UpdatedAt:        dbUser.UpdatedAt,
		EmailConfirmedAt: dbUser.EmailConfirmedAt,
		BlockedAt:        dbUser.BlockedAt,
	}
}

func toDomainVerification(dbVerification sqlc.Verification) *domain.Verification {
	return &domain.Verification{
		ID:        dbVerification.ID,
		Flow:      domain.VerificationFlow(dbVerification.Flow),
		Token:     dbVerification.Token,
		CreatedAt: dbVerification.CreatedAt,
		ExpiresAt: dbVerification.ExpiresAt,
		Payload:   dbVerification.Payload,
		UserID:    dbVerification.UserID,
	}
}

func toDomainSession(dbSession sqlc.Session) *domain.Session {
	return &domain.Session{
		ID:         dbSession.ID,
		UserID:     dbSession.UserID,
		Token:      dbSession.Token,
		IPAddress:  dbSession.IpAddress,
		UserAgent:  dbSession.UserAgent,
		DeviceName: dbSession.DeviceName,
		ExpiresAt:  dbSession.ExpiresAt,
		CreatedAt:  dbSession.CreatedAt,
	}
}

func toCreateUserParams(user *domain.User) sqlc.CreateUserParams {
	var passwordHash string
	if user.PasswordHash != nil {
		passwordHash = *user.PasswordHash
	}

	return sqlc.CreateUserParams{
		ID:               user.ID,
		Name:             user.Name,
		Email:            user.Email,
		Status:           string(user.Status),
		PasswordHash:     passwordHash,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
		EmailConfirmedAt: user.EmailConfirmedAt,
		BlockedAt:        user.BlockedAt,
	}
}

func toCreateVerificationParams(verification *domain.Verification) sqlc.CreateVerificationParams {
	return sqlc.CreateVerificationParams{
		ID:        verification.ID,
		Flow:      string(verification.Flow),
		Token:     verification.Token,
		CreatedAt: verification.CreatedAt,
		ExpiresAt: verification.ExpiresAt,
		Payload:   verification.Payload,
		UserID:    verification.UserID,
	}
}

func toCreateSessionParams(session *domain.Session) sqlc.CreateSessionParams {
	return sqlc.CreateSessionParams{
		ID:         session.ID,
		UserID:     session.UserID,
		Token:      session.Token,
		IpAddress:  session.IPAddress,
		UserAgent:  session.UserAgent,
		DeviceName: session.DeviceName,
		ExpiresAt:  session.ExpiresAt,
		CreatedAt:  session.CreatedAt,
	}
}
