package biz

import (
	"budeku/iam/internal/enum"
	"context"
	"time"

	"github.com/google/uuid"
)

type SessionCommand struct {
	Identify        uuid.UUID
	AppID           int
	UserID          int
	IP              string
	UserAgent       string
	AuthMedhod      enum.SessionAuthMethod
	AuthenticatedAt time.Time
	ExpiresAt       time.Time
}

type SessionCreateResp struct {
	ID int
}

type SessionStore interface {
	Create(ctx context.Context, body SessionCommand) (*SessionCreateResp, error)

	UpdateLastActive(ctx context.Context, id int, t time.Time) error
}

type SessionBiz struct {
	store SessionStore
}

func NewSessionBiz(store SessionStore) *SessionBiz {
	return &SessionBiz{store: store}
}
