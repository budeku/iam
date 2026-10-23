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
	ID              int                    `json:"id"`
	Identity        uuid.UUID              `json:"identity"`
	AppID           int                    `json:"app_id"`
	UserID          int                    `json:"user_id"`
	IP              string                 `json:"ip"`
	UserAgent       string                 `json:"user_agent"`
	AuthMethod      enum.SessionAuthMethod `json:"auth_method"`
	AuthenticatedAt time.Time              `json:"authenticated_at"`
	LastActiveAt    time.Time              `json:"last_active_at"`
	RevokedAt       *time.Time             `json:"revoked_at"`
	ExpiresAt       time.Time              `json:"expires_at"`
	CreatedAt       time.Time              `json:"created_at"`

	// 关联
	User          *UserGetResp          `json:"user,omitempty"`
	RefreshTokens []RefreshTokenGetResp `json:"refresh_tokens,omitempty"`
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
