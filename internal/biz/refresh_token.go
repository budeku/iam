package biz

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RefreshTokenCommand struct {
	AppID     int
	SessionID int
	UserID    int
	Identity  []byte
	Jti       uuid.UUID
	ExpiresAt time.Time
}

type RefreshTokenCreateResp struct {
	ID int
}

type RefreshTokenStore interface {
	Create(ctx context.Context, body RefreshTokenCommand) (*RefreshTokenCreateResp, error)

	RevokedBySessionID(ctx context.Context, sessionID int) error
}

type RefreshTokenBiz struct {
	store RefreshTokenStore
}

func NewRefreshTokenBiz(store RefreshTokenStore) *RefreshTokenBiz {
	return &RefreshTokenBiz{store: store}
}
