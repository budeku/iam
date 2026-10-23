package biz

import (
	"budeku/iam/internal/enum"
	"context"
	"time"

	"github.com/google/uuid"
)

type SessionCommand struct {
	Identity        uuid.UUID
	AppID           int
	UserID          int
	IP              string
	UserAgent       string
	AuthMethod      enum.SessionAuthMethod
	AuthenticatedAt time.Time
	ExpiresAt       time.Time
}

type SessionCreateResp struct {
	ID int
}

type SessionGetResp struct {
	ID              int
	Identity        uuid.UUID
	AppID           int
	UserID          int
	IP              string
	UserAgent       string
	AuthMethod      enum.SessionAuthMethod
	AuthenticatedAt time.Time
	LastActiveAt    time.Time
	RevokedAt       *time.Time
	ExpiresAt       time.Time
	CreatedAt       time.Time
}

type SessionStore interface {
	Get(ctx context.Context, id int) (*SessionGetResp, error)

	Create(ctx context.Context, body SessionCommand) (*SessionCreateResp, error)

	UpdateLastActive(ctx context.Context, id int, t time.Time) error
	RevokedByID(ctx context.Context, id int) error
	RevokedByIdentity(ctx context.Context, identity uuid.UUID) (int, error)
	RevokeByAppID(ctx context.Context, appID int) ([]int, []uuid.UUID, error)
}

type SessionBiz struct {
	store SessionStore
}

func NewSessionBiz(store SessionStore) *SessionBiz {
	return &SessionBiz{store: store}
}
