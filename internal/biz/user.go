package biz

import (
	"budeku/iam/internal/conf"
	"budeku/iam/internal/enum"
	"budeku/iam/internal/logger"
	"budeku/iam/pkg/hashutil"
	"budeku/iam/pkg/password"
	"budeku/iam/pkg/random"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type UserGetResp struct {
	ID           int
	Username     string
	UUID         uuid.UUID
	Email        string
	Password     []byte
	Status       bool
	LastLoginAt  *time.Time
	IsSuperAdmin bool
}

type UserStore interface {
	Get(ctx context.Context, id int) (*UserGetResp, error)
	GetByUsername(ctx context.Context, username string) (*UserGetResp, error)
	GetByUUID(ctx context.Context, userUUID uuid.UUID) (*UserGetResp, error)

	UpdateLastLogin(ctx context.Context, id int, at time.Time) error
}

type JwtSinger interface {
	SignAccessToken(ctx context.Context, userUUID, sessionID uuid.UUID, name string, appID int, ttl time.Duration) (string, error)
}

type UserBiz struct {
	store             UserStore
	sessionStore      SessionStore
	refreshTokenStore RefreshTokenStore
	tx                Transactor
	signer            JwtSinger
	authToken         AuthToken
}

func NewUserBiz(
	store UserStore,
	sessionStore SessionStore,
	refreshTokenStore RefreshTokenStore,
	tx Transactor,
	signer JwtSinger,
	authToken AuthToken,
) *UserBiz {
	return &UserBiz{
		store:             store,
		sessionStore:      sessionStore,
		refreshTokenStore: refreshTokenStore,
		tx:                tx,
		signer:            signer,
		authToken:         authToken,
	}
}

type LoginResp struct {
	AccessToken  string
	RefreshToken string
}

func (b *UserBiz) Login(ctx context.Context, username, passwordStr, ip, userAgent string) (*LoginResp, error) {
	info, err := b.store.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, errors.New("账号不存在")
		}
		return nil, errors.New("验证账号密码失败")
	}

	if !info.Status {
		return nil, errors.New("账号被禁用")
	}

	ok, err := password.Validate(passwordStr, string(info.Password))
	if err != nil {
		return nil, errors.New("密码验证失败")
	}
	if !ok {
		return nil, errors.New("密码错误")
	}

	if err := b.store.UpdateLastLogin(ctx, info.ID, time.Now()); err != nil {
		logger.Error("更新登录时间错误：%s", err)
		return nil, errors.New("登录失败")
	}

	now := time.Now()
	jwtCfg := conf.Get().JWT
	// 刷新令牌和会话的过期时间分别存储，在会话有效期内可以重复刷新，会话过期后需要重新登录，刷新令牌时会话不变
	refreshTokenExpiresAt := now.Add(jwtCfg.RefreshTokenTTL)
	sessionExpiresAt := now.Add(jwtCfg.SessionTTL)

	var refreshTokenPlain string
	var sessionID int
	var sessionIdentity uuid.UUID

	err = b.tx.WithTx(ctx, func(ctx context.Context) error {
		sessionIdentity = uuid.New()

		sessionResp, err := b.sessionStore.Create(ctx, SessionCommand{
			AppID:           conf.IAMAppID,
			UserID:          info.ID,
			Identity:        sessionIdentity,
			IP:              ip,
			UserAgent:       userAgent,
			AuthMedhod:      enum.SessionAuthMethodPassword,
			AuthenticatedAt: now,
			ExpiresAt:       sessionExpiresAt,
		})
		if err != nil {
			return err
		}
		sessionID = sessionResp.ID

		refreshTokenPlain, err = b.createRefreshTokenInTx(ctx, info.ID, conf.IAMAppID, sessionID, refreshTokenExpiresAt)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error("登录创建会话事务失败：%s", err)
		return nil, errors.New("登录失败")
	}

	accessToken, err := b.signer.SignAccessToken(
		ctx,
		info.UUID,
		sessionIdentity,
		info.Username,
		conf.IAMAppID,
		jwtCfg.AccessTokenTTL,
	)
	if err != nil {
		logger.Error("签发 AccessToken 失败：%s", err)
		return nil, errors.New("登录失败")
	}

	// 令牌保存到缓存中
	if err := b.authToken.SetAccess(ctx, sessionIdentity, jwtCfg.AccessTokenTTL); err != nil {
		return nil, errors.New("登录异常")
	}
	if err := b.authToken.SetRefresh(ctx, sessionIdentity, jwtCfg.RefreshTokenTTL); err != nil {
		return nil, errors.New("登录异常")
	}

	return &LoginResp{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenPlain,
	}, nil
}

func (b *UserBiz) createRefreshTokenInTx(ctx context.Context, userID, appID, sessionID int, expiresAt time.Time) (string, error) {
	plain, err := random.Alphanumeric(64)
	if err != nil {
		return "", err
	}

	identity := hashutil.Sum256([]byte(plain)) // 数据库存储哈希值
	_, err = b.refreshTokenStore.Create(ctx, RefreshTokenCommand{
		AppID:     appID,
		UserID:    userID,
		SessionID: sessionID,
		Identity:  identity,
		Jti:       uuid.New(),
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", err
	}

	return plain, nil
}

func (b *UserBiz) Logout(ctx context.Context, sessionIdentity uuid.UUID) error {
	if err := b.authToken.DelAll(ctx, sessionIdentity); err != nil {
		logger.Error("退出清理 Redis 令牌失败：%s", err)
		return errors.New("退出异常")
	}

	return b.tx.WithTx(ctx, func(ctx context.Context) error {
		sessionInfo, err := b.sessionStore.RevokedByIdentity(ctx, sessionIdentity)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil // 会话已经不存在了
			}
			return err
		}

		if sessionInfo != nil {
			if err := b.refreshTokenStore.RevokedBySessionID(ctx, sessionInfo.ID); err != nil {
				return err
			}
		}

		return nil
	})
}

type RefreshTokenResp struct {
	AccessToken  string
	RefreshToken string
}

func (b *UserBiz) RefreshToken(ctx context.Context, oldRefreshTokenPlain string) (*RefreshTokenResp, error) {
	// 明文 Refresh Token 哈希
	oldIdentity := hashutil.Sum256([]byte(oldRefreshTokenPlain))

	var newAccessToken string
	var newRefreshTokenPlain string
	var sessionIdentity uuid.UUID

	jwtCfg := conf.Get().JWT
	now := time.Now()

	err := b.tx.WithTx(ctx, func(ctx context.Context) error {
		// 查询旧的 RefreshToken 记录
		oldRt, err := b.refreshTokenStore.GetByIdentity(ctx, oldIdentity)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return errors.New("无效的刷新令牌")
			}
			return err
		}

		// 检查旧的 Refresh Token 是否已过期
		if now.After(oldRt.ExpiresAt) {
			return errors.New("刷新令牌已过期，请重新登录")
		}

		// 检查旧的 Refresh Token 是否已经被吊销
		if oldRt.RevokedAt != nil {
			logger.Warn("检测到 RefreshToken 重放攻击嫌疑，正在强制吊销 Session ID: %d", oldRt.SessionID)
			_ = b.sessionStore.RevokedByID(ctx, oldRt.SessionID)
			_ = b.refreshTokenStore.RevokedBySessionID(ctx, oldRt.SessionID)
			return errors.New("检测到安全风险，会话已被强制终止，请重新登录")
		}

		// 查询关联的 Session 信息
		sessionInfo, err := b.sessionStore.GetByID(ctx, oldRt.SessionID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return errors.New("关联的会话不存在")
			}
			return err
		}

		// 检查 Session 是否已被吊销
		if sessionInfo.RevokedAt != nil {
			return errors.New("会话已被吊销，请重新登录")
		}

		// 检查 Session 是否超过了绝对生命周期上限 (SessionTTL / ExpiresAt)
		if now.After(sessionInfo.ExpiresAt) {
			return errors.New("会话已过绝对生命周期，请重新登录")
		}

		// 用户信息
		userInfo, err := b.store.Get(ctx, sessionInfo.UserID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return errors.New("关联的用户不存在或已被删除")
			}
			return err
		}

		// 如果用户被禁用了，在刷新令牌时也要拒绝
		if !userInfo.Status {
			return errors.New("账号已被禁用")
		}

		// 吊销当前的旧 Refresh Token (用过即作废)
		if err := b.refreshTokenStore.RevokeByID(ctx, oldRt.ID); err != nil {
			return err
		}

		// 更新 Session 的最后活跃时间
		if err := b.sessionStore.UpdateLastActive(ctx, sessionInfo.ID, now); err != nil {
			return err
		}

		// 签发新的 Refresh Token，绑定到同一个 Session
		refreshTokenExpiresAt := now.Add(jwtCfg.RefreshTokenTTL)
		if refreshTokenExpiresAt.After(sessionInfo.ExpiresAt) {
			refreshTokenExpiresAt = sessionInfo.ExpiresAt // 封顶不超过 Session 绝对上限
		}

		sessionIdentity = sessionInfo.Identity
		newRefreshTokenPlain, err = b.createRefreshTokenInTx(ctx, sessionInfo.UserID, sessionInfo.AppID, sessionInfo.ID, refreshTokenExpiresAt)
		if err != nil {
			return err
		}

		// 签发新的 Access Token
		newAccessToken, err = b.signer.SignAccessToken(
			ctx,
			userInfo.UUID,
			sessionIdentity,
			userInfo.Username,
			sessionInfo.AppID,
			jwtCfg.AccessTokenTTL,
		)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error("刷新令牌事务失败：%s", err)
		return nil, err
	}

	// 先删除再写入
	if err := b.authToken.DelAll(ctx, sessionIdentity); err != nil {
		logger.Error("清理 Redis 旧令牌缓存失败：%s", err)
	}

	if err := b.authToken.SetAccess(ctx, sessionIdentity, jwtCfg.AccessTokenTTL); err != nil {
		logger.Error("设置 Redis AccessToken 缓存失败：%s", err)
	}
	if err := b.authToken.SetRefresh(ctx, sessionIdentity, jwtCfg.RefreshTokenTTL); err != nil {
		logger.Error("设置 Redis RefreshToken 缓存失败：%s", err)
	}

	return &RefreshTokenResp{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshTokenPlain,
	}, nil
}

type UserMeResp struct {
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	UUID         uuid.UUID `json:"uuid"`
	IsSuperAdmin bool      `json:"is_super_admin"`
}

func (b *UserBiz) Me(ctx context.Context, userUUID uuid.UUID) (*UserMeResp, error) {
	u, err := b.store.GetByUUID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
	}

	return &UserMeResp{
		Username:     u.Username,
		Email:        u.Email,
		UUID:         u.UUID,
		IsSuperAdmin: u.IsSuperAdmin,
	}, nil
}

func (b *UserBiz) GetByUUID(ctx context.Context, userUUID uuid.UUID) (*UserGetResp, error) {
	info, err := b.store.GetByUUID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		logger.Error("%v", err)
		return nil, err
	}
	return info, nil
}
