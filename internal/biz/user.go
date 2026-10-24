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
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	UUID         uuid.UUID  `json:"uuid"`
	Email        string     `json:"email"`
	Password     []byte     `json:"password"`
	Status       bool       `json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	IsSuperAdmin bool       `json:"is_super_admin"`
}

type UserInfoResp struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	UUID         uuid.UUID  `json:"uuid"`
	Email        string     `json:"email"`
	Status       bool       `json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	IsSuperAdmin bool       `json:"is_super_admin"`
}

type UserListItemResp struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	UUID         uuid.UUID  `json:"uuid"`
	Email        string     `json:"email"`
	Status       bool       `json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	IsSuperAdmin bool       `json:"is_super_admin"`
}

type UserCreateResp struct {
	ID int `json:"id"`
}

type UserSessionResp struct {
	ID              int                           `json:"id"`
	Identity        uuid.UUID                     `json:"identity"`
	AppID           int                           `json:"app_id"`
	AppName         string                        `json:"app_name"`
	IP              string                        `json:"ip"`
	UserAgent       string                        `json:"user_agent"`
	AuthMethod      enum.SessionAuthMethod        `json:"auth_method"`
	AuthenticatedAt time.Time                     `json:"authenticated_at"`
	LastActiveAt    time.Time                     `json:"last_active_at"`
	ExpiresAt       time.Time                     `json:"expires_at"`
	RefreshTokens   []UserSessionRefreshTokenResp `json:"refresh_tokens"`
	IsCurrent       bool                          `json:"is_current"`
}

type UserSessionRefreshTokenResp struct {
	ID        int       `json:"id"`
	Jti       uuid.UUID `json:"jti"`
	ExpiresAt time.Time `json:"expires_at"`
}

type UserStore interface {
	List(ctx context.Context, keyword string, page, perPage int) ([]UserListItemResp, int, error)

	Get(ctx context.Context, id int) (*UserGetResp, error)
	GetByUsername(ctx context.Context, username string) (*UserGetResp, error)
	GetByUUID(ctx context.Context, userUUID uuid.UUID) (*UserGetResp, error)

	Duplicate(ctx context.Context, username, email string, id ...int) (bool, error)

	Create(ctx context.Context, username, email, passwordHash string, uuid uuid.UUID, status bool) (*UserCreateResp, error)

	Update(ctx context.Context, id int, username, email, passwordHash string, status bool) (*UserGetResp, error)
	UpdateLastLogin(ctx context.Context, id int, at time.Time) error
	UpdateStatus(ctx context.Context, id int, status bool) error
	UpdatePassword(ctx context.Context, id int, passwordHash string) error

	Delete(ctx context.Context, id int) error
}

type JwtSinger interface {
	SignAccessToken(ctx context.Context, userUUID, sessionID uuid.UUID, name string, appID int, ttl time.Duration, ip, userAgent string) (string, error)
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

	now := time.Now()
	jwtCfg := conf.Get().JWT
	// 刷新令牌和会话的过期时间分别存储，在会话有效期内可以重复刷新，会话过期后需要重新登录，刷新令牌时会话不变
	refreshTokenExpiresAt := now.Add(jwtCfg.RefreshTokenTTL)
	sessionExpiresAt := now.Add(jwtCfg.SessionTTL)

	var refreshTokenPlain string
	var accessToken string
	sessionIdentity := uuid.New()

	// 更新最后登录时间、创建会话、创建 RefreshToken、签发 AccessToken
	err = b.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := b.store.UpdateLastLogin(ctx, info.ID, now); err != nil {
			return err
		}

		sessionResp, err := b.sessionStore.Create(ctx, SessionCreateCommand{
			AppID:           conf.IAMAppID,
			UserID:          info.ID,
			Identity:        sessionIdentity,
			IP:              ip,
			UserAgent:       userAgent,
			AuthMethod:      enum.SessionAuthMethodPassword,
			AuthenticatedAt: now,
			ExpiresAt:       sessionExpiresAt,
		})
		if err != nil {
			return err
		}
		sessionID := sessionResp.ID

		refreshTokenPlain, err = b.createRefreshTokenInTx(ctx, info.ID, conf.IAMAppID, sessionID, refreshTokenExpiresAt)
		if err != nil {
			return err
		}

		// 生成时也会操作数据库，这里放在事务中
		accessToken, err = b.signer.SignAccessToken(
			ctx,
			info.UUID,
			sessionIdentity,
			info.Username,
			conf.IAMAppID,
			jwtCfg.AccessTokenTTL,
			ip,
			userAgent,
		)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error("登录事务执行失败：%s", err)
		return nil, errors.New("登录失败")
	}

	// 令牌保存到缓存中
	if err := b.authToken.SetAccess(ctx, sessionIdentity, jwtCfg.AccessTokenTTL); err != nil {
		logger.Error("设置 Redis AccessToken 缓存失败：%s", err)
		return nil, errors.New("登录异常")
	}
	if err := b.authToken.SetRefresh(ctx, sessionIdentity, jwtCfg.RefreshTokenTTL); err != nil {
		logger.Error("设置 Redis RefreshToken 缓存失败：%s", err)
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
	_, err = b.refreshTokenStore.Create(ctx, RefreshTokenCreateCommand{
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
		sessionID, err := b.sessionStore.RevokedByIdentity(ctx, sessionIdentity)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil // 会话已经不存在了
			}
			return err
		}

		if sessionID != 0 {
			if err := b.refreshTokenStore.RevokedBySessionID(ctx, sessionID); err != nil {
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
		oldRt, err := b.refreshTokenStore.GetByIdentityWithEdges(ctx, oldIdentity)
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

		sessionInfo := oldRt.Session
		if sessionInfo == nil {
			return errors.New("关联的会话不存在")
		}

		// 检查 Session 是否已被吊销
		if sessionInfo.RevokedAt != nil {
			return errors.New("会话已被吊销，请重新登录")
		}

		// 检查 Session 是否超过了绝对生命周期上限
		if now.After(sessionInfo.ExpiresAt) {
			return errors.New("会话已过绝对生命周期，请重新登录")
		}

		userInfo := sessionInfo.User
		if userInfo == nil {
			return errors.New("关联的用户不存在或已被删除")
		}

		// 如果用户被禁用了，拒绝刷新
		if !userInfo.Status {
			return errors.New("账号已被禁用")
		}

		// 吊销当前的旧 Refresh Token
		if err := b.refreshTokenStore.RevokeByID(ctx, oldRt.ID); err != nil {
			return err
		}

		// 更新 Session 最后活跃时间
		if err := b.sessionStore.UpdateLastActive(ctx, sessionInfo.ID, now); err != nil {
			return err
		}

		// 签发新的 Refresh Token，绑定到同一个 Session
		refreshTokenExpiresAt := now.Add(jwtCfg.RefreshTokenTTL)
		if refreshTokenExpiresAt.After(sessionInfo.ExpiresAt) {
			refreshTokenExpiresAt = sessionInfo.ExpiresAt // 封顶不超过 Session 绝对上限
		}

		sessionIdentity = sessionInfo.Identity
		newRefreshTokenPlain, err = b.createRefreshTokenInTx(ctx, userInfo.ID, sessionInfo.AppID, sessionInfo.ID, refreshTokenExpiresAt)
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
			sessionInfo.IP,
			sessionInfo.UserAgent,
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
		logger.Error("获取当前用户信息失败: %v", err)
		return nil, err
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

func (b *UserBiz) List(ctx context.Context, keyword string, page, perPage int) ([]UserListItemResp, int, error) {
	return b.store.List(ctx, keyword, page, perPage)
}

func (b *UserBiz) Info(ctx context.Context, id int) (*UserInfoResp, error) {
	info, err := b.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return &UserInfoResp{
		ID:           info.ID,
		Username:     info.Username,
		Email:        info.Email,
		UUID:         info.UUID,
		Status:       info.Status,
		LastLoginAt:  info.LastLoginAt,
		IsSuperAdmin: info.IsSuperAdmin,
	}, nil
}

func (b *UserBiz) Create(ctx context.Context, username, email string, passwordStr string) (*UserCreateResp, error) {
	exist, err := b.store.Duplicate(ctx, username, email)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, ErrAlreadyExists
	}

	passwordHash, err := password.Hash(passwordStr)
	if err != nil {
		return nil, err
	}

	return b.store.Create(ctx, username, email, passwordHash, uuid.New(), true)
}

func (b *UserBiz) Update(ctx context.Context, id int, username, email string, passwordStr string, status bool) error {
	oldUser, err := b.store.Get(ctx, id)
	if err != nil {
		return err
	}

	if oldUser.IsSuperAdmin && !status {
		return errors.New("超级管理员账号不能被禁用")
	}

	exist, err := b.store.Duplicate(ctx, username, email, id)
	if err != nil {
		return err
	}
	if exist {
		return ErrAlreadyExists
	}

	var passwordHash string
	if passwordStr != "" {
		var err error
		passwordHash, err = password.Hash(passwordStr)
		if err != nil {
			return err
		}
	}

	// 修改任意一项则重新登录
	needRevoke := (oldUser.Status && !status) ||
		(passwordStr != "") ||
		(oldUser.Username != username) ||
		(oldUser.Email != email)

	err = b.tx.WithTx(ctx, func(txCtx context.Context) error {
		_, err := b.store.Update(txCtx, id, username, email, passwordHash, status)
		return err
	})

	if err != nil {
		return err
	}

	if needRevoke {
		go b.revokeUserSessionsAndTokens(id)
	}

	return nil
}

func (b *UserBiz) UpdateStatus(ctx context.Context, id int, status bool) error {
	oldUser, err := b.store.Get(ctx, id)
	if err != nil {
		return err
	}

	if oldUser.Status == status {
		return nil
	}

	if oldUser.IsSuperAdmin && !status {
		return errors.New("超级管理员账号不能被禁用")
	}

	needRevoke := oldUser.Status && !status

	err = b.tx.WithTx(ctx, func(txCtx context.Context) error {
		return b.store.UpdateStatus(txCtx, id, status)
	})

	if err != nil {
		return err
	}

	if needRevoke {
		go b.revokeUserSessionsAndTokens(id)
	}

	return nil
}

func (b *UserBiz) UpdatePassword(ctx context.Context, userID int, oldPasswordStr, newPasswordStr string) error {
	if oldPasswordStr == "" || newPasswordStr == "" {
		return errors.New("密码不能为空")
	}
	if oldPasswordStr == newPasswordStr {
		return errors.New("新密码不能与旧密码相同")
	}

	info, err := b.store.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	ok, err := password.Validate(oldPasswordStr, string(info.Password))
	if err != nil {
		return errors.New("密码验证失败")
	}
	if !ok {
		return errors.New("旧密码错误")
	}

	newPasswordHash, err := password.Hash(newPasswordStr)
	if err != nil {
		logger.Error("新密码加密失败: %v", err)
		return errors.New("密码处理失败")
	}

	err = b.tx.WithTx(ctx, func(txCtx context.Context) error {
		return b.store.UpdatePassword(txCtx, userID, newPasswordHash)
	})

	if err != nil {
		logger.Error("修改密码事务执行失败 (UserID: %d): %v", userID, err)
		return errors.New("修改密码失败")
	}

	go b.revokeUserSessionsAndTokens(userID)

	return nil
}

func (b *UserBiz) Delete(ctx context.Context, id int) error {
	oldUser, err := b.store.Get(ctx, id)
	if err != nil {
		return err
	}

	if oldUser.IsSuperAdmin {
		return errors.New("超级管理员账号不能被删除")
	}

	err = b.tx.WithTx(ctx, func(txCtx context.Context) error {
		return b.store.Delete(txCtx, id)
	})

	if err != nil {
		return err
	}

	go b.revokeUserSessionsAndTokens(id)

	return nil
}

func (b *UserBiz) revokeUserSessionsAndTokens(userID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var sessionIdentities []uuid.UUID

	err := b.tx.WithTx(ctx, func(txCtx context.Context) error {
		sessionIDs, identities, err := b.sessionStore.RevokeByUserID(txCtx, userID)
		if err != nil {
			return err
		}
		sessionIdentities = identities

		if len(sessionIDs) > 0 {
			if err := b.refreshTokenStore.RevokedBySessionIDs(txCtx, sessionIDs); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		logger.Error("级联吊销用户(ID: %d)的会话和令牌数据库操作失败: %v", userID, err)
		return
	}

	if len(sessionIdentities) > 0 {
		if err := b.authToken.DelAllBySessionIdentities(context.Background(), sessionIdentities); err != nil {
			logger.Error("清理用户(ID: %d)关联的 Redis 凭证失败: %v", userID, err)
		}
	}
}

func (b *UserBiz) Session(ctx context.Context, id int, currentSessionID uuid.UUID) ([]UserSessionResp, error) {
	return b.sessionStore.GetUserSessions(ctx, id, currentSessionID)
}

func (b *UserBiz) RevokeSession(ctx context.Context, currentUserID int, sessionID int) error {
	sessionInfo, err := b.sessionStore.Get(ctx, sessionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return errors.New("会话不存在")
		}
		return err
	}

	if sessionInfo.UserID != currentUserID {
		return errors.New("无权操作该会话")
	}

	// 如果会话已经被吊销了，直接返回成功
	if sessionInfo.RevokedAt != nil {
		return nil
	}

	sessionIdentity := sessionInfo.Identity

	err = b.tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := b.sessionStore.RevokedByID(txCtx, sessionID); err != nil {
			return err
		}
		if err := b.refreshTokenStore.RevokedBySessionID(txCtx, sessionID); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		logger.Error("下线会话数据库操作失败 (SessionID: %d): %v", sessionID, err)
		return errors.New("下线会话失败")
	}

	if err := b.authToken.DelAll(ctx, sessionIdentity); err != nil {
		logger.Error("清理 Redis 令牌缓存失败 (Identity: %s): %v", sessionIdentity, err)
	}

	return nil
}
