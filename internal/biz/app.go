package biz

import (
	"budeku/iam/internal/logger"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type AppListItemResp struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	ClientID  string    `json:"client_id"`
	Status    bool      `json:"status"`
	IsSystem  bool      `json:"is_system"`
	CreatedAt time.Time `json:"created_at"`
}

type AppGetResp struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	ClientID     string    `json:"client_id"`
	RedirectUris []string  `json:"redirect_uris"`
	Status       bool      `json:"status"`
	IsSystem     bool      `json:"is_system"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AppStore interface {
	List(ctx context.Context, keyword string, page, perPage int) ([]AppListItemResp, int, error)
	ListByIds(ctx context.Context, ids []int) ([]AppListItemResp, error)

	Get(ctx context.Context, id int) (*AppGetResp, error)
	GetByClientID(ctx context.Context, clientID string) (*AppGetResp, error)

	Duplicate(ctx context.Context, name, clientID string, id ...int) (bool, error)

	Create(ctx context.Context, name, clientID string, redirectUris []string, status bool) (*AppGetResp, error)

	Update(ctx context.Context, id int, name, clientID string, redirectUris []string, status bool) (*AppGetResp, error)
	UpdateStatus(ctx context.Context, id int, status bool) error

	Delete(ctx context.Context, id int) error
}

type AppBiz struct {
	store             AppStore
	sessionStore      SessionStore
	refreshTokenStore RefreshTokenStore
	transactor        Transactor
	authToken         AuthToken
}

func NewAppBiz(
	store AppStore,
	sessionStore SessionStore,
	refreshTokenStore RefreshTokenStore,
	transactor Transactor,
	authToken AuthToken,
) *AppBiz {
	return &AppBiz{
		store:             store,
		sessionStore:      sessionStore,
		refreshTokenStore: refreshTokenStore,
		transactor:        transactor,
		authToken:         authToken,
	}
}

func (b *AppBiz) List(ctx context.Context, keyword string, page, perPage int) ([]AppListItemResp, int, error) {
	return b.store.List(ctx, keyword, page, perPage)
}

func (b *AppBiz) Info(ctx context.Context, id int) (*AppGetResp, error) {
	return b.store.Get(ctx, id)
}

func (b *AppBiz) Create(ctx context.Context, name, clientID string, redirectUris []string, status bool) (*AppGetResp, error) {
	exist, err := b.store.Duplicate(ctx, name, clientID)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, ErrAlreadyExists
	}
	return b.store.Create(ctx, name, clientID, redirectUris, status)
}

func (b *AppBiz) Update(ctx context.Context, id int, name, clientID string, redirectUris []string, status bool) (*AppGetResp, error) {
	oldApp, err := b.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	exist, err := b.store.Duplicate(ctx, name, clientID, id)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errors.New("应用名称或 ClientID 已存在")
	}

	var updated *AppGetResp
	needRevoke := oldApp.Status && !status

	err = b.transactor.WithTx(ctx, func(txCtx context.Context) error {
		var innerErr error
		updated, innerErr = b.store.Update(txCtx, id, name, clientID, redirectUris, status)
		return innerErr
	})

	if err != nil {
		return nil, err
	}

	if needRevoke {
		go b.revokeAppSessionsAndTokens(id)
	}

	return updated, nil
}

func (b *AppBiz) UpdateStatus(ctx context.Context, id int, status bool) error {
	oldApp, err := b.store.Get(ctx, id)
	if err != nil {
		return err
	}

	if oldApp.Status == status {
		return nil
	}

	needRevoke := oldApp.Status && !status

	err = b.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return b.store.UpdateStatus(txCtx, id, status)
	})

	if err != nil {
		return err
	}

	if needRevoke {
		go b.revokeAppSessionsAndTokens(id)
	}

	return nil
}

func (b *AppBiz) Delete(ctx context.Context, id int) error {
	if _, err := b.store.Get(ctx, id); err != nil {
		return err
	}

	err := b.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return b.store.Delete(txCtx, id)
	})

	if err != nil {
		return err
	}

	go b.revokeAppSessionsAndTokens(id)

	return nil
}

func (b *AppBiz) revokeAppSessionsAndTokens(appID int) {
	// 异步执行，创建一个新的context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var sessionIdentities []uuid.UUID

	err := b.transactor.WithTx(ctx, func(txCtx context.Context) error {
		sessionIDs, identities, err := b.sessionStore.RevokeByAppID(txCtx, appID)
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
		logger.Error("级联吊销应用(ID: %d)的会话和令牌数据库操作失败: %v", appID, err)
		return
	}

	if len(sessionIdentities) > 0 {
		if err := b.authToken.DelAllBySessionIdentities(context.Background(), sessionIdentities); err != nil {
			logger.Error("清理应用(ID: %d)关联的 Redis 凭证失败: %v", appID, err)
		}
	}
}
