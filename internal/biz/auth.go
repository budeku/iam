package biz

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AuthToken interface {
	SetAccess(ctx context.Context, sessionID uuid.UUID, ttl time.Duration) error
	ExistsAccess(ctx context.Context, sessionID uuid.UUID) (bool, error)
	DelAccess(ctx context.Context, sessionID uuid.UUID) error
	SetRefresh(ctx context.Context, sessionID uuid.UUID, ttl time.Duration) error
	ExistsRefresh(ctx context.Context, sessionID uuid.UUID) (bool, error)
	DelRefresh(ctx context.Context, sessionID uuid.UUID) error
	DelAll(ctx context.Context, sessionID uuid.UUID) error
}
