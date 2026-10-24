package biz

import (
	"budeku/iam/internal/consts"
	"budeku/iam/internal/enum"
	"budeku/iam/internal/infra/redis"
	"budeku/iam/internal/logger"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SessionCreateCommand struct {
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
	GetUserSessions(ctx context.Context, userID int, currentSessionID uuid.UUID) ([]UserSessionResp, error)

	Create(ctx context.Context, body SessionCreateCommand) (*SessionCreateResp, error)

	UpdateLastActive(ctx context.Context, id int, t time.Time) error
	UpdateLastActiveByIdentity(ctx context.Context, identity uuid.UUID, t time.Time) error

	RevokedByID(ctx context.Context, id int) error
	RevokedByIdentity(ctx context.Context, identity uuid.UUID) (int, error)
	RevokeByAppID(ctx context.Context, appID int) ([]int, []uuid.UUID, error)
	RevokeByUserID(ctx context.Context, userID int) ([]int, []uuid.UUID, error)
}

type SessionBiz struct {
	store SessionStore
	rdb   *redis.Redis
}

func NewSessionBiz(store SessionStore) *SessionBiz {
	return &SessionBiz{
		store: store,
		rdb:   redis.DB(),
	}
}

func (b *SessionBiz) touchCacheKey(identity uuid.UUID) string {
	return fmt.Sprintf("%s:%s", consts.RedisSessionTouch, identity.String())
}

func (b *SessionBiz) Touch(ctx context.Context, identity uuid.UUID) {
	cacheKey := b.touchCacheKey(identity)

	// 利用 SetNX 设置 5 分钟的冷却标记
	success, err := b.rdb.SetNX(ctx, cacheKey, "1", 5*time.Minute)
	if err != nil {
		logger.Error("会话活跃度 Redis SetNX 失败, 会话=%s err=%v", identity, err)
		return
	}

	if !success {
		return
	}

	t := time.Now()
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := b.store.UpdateLastActiveByIdentity(bgCtx, identity, t); err != nil {
			logger.Error("更新会话活动失败, 会话=%s err=%v", identity, err)

			// 数据库更新失败，删除缓存
			_ = b.rdb.Del(bgCtx, cacheKey)
		}
	}()
}
