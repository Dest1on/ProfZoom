package auth

import (
	"context"

	"profzom/internal/common"
)

type OTPRepository interface {
	UpsertCode(ctx context.Context, userID, code string, expiresAtUnix int64, attemptsLeft int) error
	VerifyCode(ctx context.Context, userID, code string, nowUnix int64) (bool, error)
	GetState(ctx context.Context, userID string) (*OTPState, error)
	InvalidateCode(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context, beforeUnix int64) error
}

type OTPState struct {
	UserID       string
	AttemptsLeft int
	ExpiresAt    int64
	RequestedAt  int64
}

type RefreshTokenRepository interface {
	Store(ctx context.Context, token RefreshToken) error
	GetByToken(ctx context.Context, token string) (*RefreshToken, error)
	Revoke(ctx context.Context, token string, revokedAtUnix int64) error
	RevokeAll(ctx context.Context, userID common.UUID, revokedAtUnix int64) error
}
