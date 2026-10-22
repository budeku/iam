package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/consts"
	"budeku/iam/internal/infra/redis"
	"context"
	"time"

	"github.com/google/uuid"
)

var _ biz.AuthToken = (*AuthToken)(nil)

type AuthToken struct {
	rdb *redis.Redis
}

func NewAuthToken() *AuthToken {
	return &AuthToken{rdb: redis.DB()}
}

func (t *AuthToken) accessKey(sessionID uuid.UUID) string {
	return consts.RedisAccessToken + ":" + sessionID.String()
}

func (t *AuthToken) SetAccess(ctx context.Context, sessionID uuid.UUID, ttl time.Duration) error {
	return t.rdb.Set(ctx, t.accessKey(sessionID), "1", ttl)
}

func (t *AuthToken) ExistsAccess(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	return t.rdb.Exists(ctx, t.accessKey(sessionID))
}

func (t *AuthToken) DelAccess(ctx context.Context, sessionID uuid.UUID) error {
	return t.rdb.Del(ctx, t.accessKey(sessionID))
}

func (t *AuthToken) refreshKey(sessionID uuid.UUID) string {
	return consts.RedisRefreshToken + ":" + sessionID.String()
}

func (t *AuthToken) SetRefresh(ctx context.Context, sessionID uuid.UUID, ttl time.Duration) error {
	return t.rdb.Set(ctx, t.refreshKey(sessionID), "1", ttl)
}

func (t *AuthToken) ExistsRefresh(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	return t.rdb.Exists(ctx, t.refreshKey(sessionID))
}

func (t *AuthToken) DelRefresh(ctx context.Context, sessionID uuid.UUID) error {
	return t.rdb.Del(ctx, t.refreshKey(sessionID))
}

func (t *AuthToken) DelAll(ctx context.Context, sessionID uuid.UUID) error {
	return t.rdb.Del(ctx, t.accessKey(sessionID), t.refreshKey(sessionID))
}

func (t *AuthToken) DelAllBySessionIdentities(ctx context.Context, identities []uuid.UUID) error {
	if len(identities) == 0 {
		return nil
	}

	keys := make([]string, 0, len(identities)*2)
	for _, id := range identities {
		keys = append(keys, t.accessKey(id), t.refreshKey(id))
	}

	return t.rdb.Del(ctx, keys...)
}
