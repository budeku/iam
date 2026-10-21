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
	GetByUsername(ctx context.Context, username string) (*UserGetResp, error)

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
}

func NewUserBiz(
	store UserStore,
	sessionStore SessionStore,
	refreshTokenStore RefreshTokenStore,
	tx Transactor,
	signer JwtSinger,
) *UserBiz {
	return &UserBiz{
		store:             store,
		sessionStore:      sessionStore,
		refreshTokenStore: refreshTokenStore,
		tx:                tx,
		signer:            signer,
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
	refreshTokenTTL := conf.Get().JWT.RefreshTokenTTL
	expiresAt := now.Add(refreshTokenTTL)

	var refreshTokenPlain string
	var sessionID int
	var sessionIdentity uuid.UUID

	err = b.tx.WithTx(ctx, func(ctx context.Context) error {
		sessionIdentity = uuid.New()

		sessionResp, err := b.sessionStore.Create(ctx, SessionCommand{
			AppID:           conf.IAMAppID,
			UserID:          info.ID,
			Identify:        sessionIdentity,
			IP:              ip,
			UserAgent:       userAgent,
			AuthMedhod:      enum.SessionAuthMethodPassword,
			AuthenticatedAt: now,
			ExpiresAt:       expiresAt,
		})
		if err != nil {
			return err
		}
		sessionID = sessionResp.ID

		refreshTokenPlain, err = b.createRefreshTokenInTx(ctx, info.ID, conf.IAMAppID, sessionID, expiresAt)
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
		conf.Get().JWT.AccessTokenTTL,
	)
	if err != nil {
		logger.Error("签发 AccessToken 失败：%s", err)
		return nil, errors.New("登录失败")
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
		Identify:  identity,
		Jti:       uuid.New(),
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", err
	}

	return plain, nil
}
