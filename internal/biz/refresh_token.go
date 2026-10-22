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

type RefreshTokenInfoResp struct {
	ID        int
	AppID     int
	SessionID int
	UserID    int
	Identity  []byte
	Jti       uuid.UUID
	RevokedAt *time.Time
	ExpiresAt time.Time
	CreatedAt time.Time
}

type RefreshTokenStore interface {
	GetByIdentity(ctx context.Context, identity []byte) (*RefreshTokenInfoResp, error)

	Create(ctx context.Context, body RefreshTokenCommand) (*RefreshTokenCreateResp, error)

	RevokeByID(ctx context.Context, id int) error
	RevokedBySessionID(ctx context.Context, sessionID int) error
	RevokedBySessionIDs(ctx context.Context, sessionIDs []int) error
}

type RefreshTokenBiz struct {
	store RefreshTokenStore
}

func NewRefreshTokenBiz(store RefreshTokenStore) *RefreshTokenBiz {
	return &RefreshTokenBiz{store: store}
}
