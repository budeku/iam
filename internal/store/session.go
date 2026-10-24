package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/infra/ent/db/app"
	"budeku/iam/internal/infra/ent/db/refreshtoken"
	"budeku/iam/internal/infra/ent/db/session"
	"budeku/iam/internal/infra/ent/db/user"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var _ biz.SessionStore = (*SessionStore)(nil)

type SessionStore struct {
	*baseStore
}

func NewSessionStore(client *db.Client) *SessionStore {
	return &SessionStore{newBaseStore(client)}
}

func (s *SessionStore) Get(ctx context.Context, id int) (*biz.SessionGetResp, error) {
	info, err := s.Client(ctx).Session.Query().
		Where(session.IDEQ(id)).
		WithUser(func(query *db.UserQuery) {
			query.Select(user.FieldID, user.FieldUsername, user.FieldEmail)
		}).
		WithRefreshTokens(func(query *db.RefreshTokenQuery) {
			query.Where(refreshtoken.RevokedAtIsNil())
		}).
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, biz.ErrNotFound
		}
		return nil, err
	}

	return toSessionGet(info), nil
}

func (s *SessionStore) GetActiveSessionsByUser(ctx context.Context, userID int, appID int) ([]*db.Session, error) {
	return s.Client(ctx).Session.Query().
		Where(
			session.UserIDEQ(userID),
			session.AppIDEQ(appID),
			session.RevokedAtIsNil(),
			session.HasUserWith(user.StatusEQ(true)),
		).
		All(ctx)
}

func (s *SessionStore) GetUserSessions(ctx context.Context, userID int, currentSessionID uuid.UUID) ([]biz.UserSessionResp, error) {
	list, err := s.client.Session.Query().
		Where(
			session.UserID(userID),
			session.RevokedAtIsNil(),
			session.ExpiresAtGT(time.Now()),
		).
		WithRefreshTokens(func(query *db.RefreshTokenQuery) {
			query.Where(
				refreshtoken.RevokedAtIsNil(),
				refreshtoken.ExpiresAtGT(time.Now()),
			)
		}).
		All(ctx)

	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return []biz.UserSessionResp{}, nil
	}

	appIDMap := make(map[int]struct{})
	for _, s := range list {
		if s.AppID > 0 {
			appIDMap[s.AppID] = struct{}{}
		}
	}

	var appIDs []int
	for id := range appIDMap {
		appIDs = append(appIDs, id)
	}

	appNameMap := make(map[int]string)
	if len(appIDs) > 0 {
		apps, err := s.client.App.Query().
			Where(app.IDIn(appIDs...)).
			Select(app.FieldID, app.FieldName).
			All(ctx)
		if err != nil {
			return nil, err
		}
		for _, a := range apps {
			appNameMap[a.ID] = a.Name
		}
	}

	return toUserSessions(list, currentSessionID, appNameMap), nil
}

func (s *SessionStore) Create(ctx context.Context, body biz.SessionCreateCommand) (*biz.SessionCreateResp, error) {
	info, err := s.Client(ctx).Session.Create().
		SetIdentity(body.Identity).
		SetAppID(body.AppID).
		SetUserID(body.UserID).
		SetIP(body.IP).
		SetUserAgent(body.UserAgent).
		SetAuthMethod(body.AuthMethod).
		SetAuthenticatedAt(body.AuthenticatedAt).
		SetExpiresAt(body.ExpiresAt).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return toSessionCreate(info), nil
}

func (s *SessionStore) UpdateLastActive(ctx context.Context, id int, t time.Time) error {
	err := s.Client(ctx).Session.UpdateOneID(id).
		SetLastActiveAt(t).
		Exec(ctx)

	if err != nil {
		if db.IsNotFound(err) {
			return biz.ErrNotFound
		}
		return err
	}

	return nil
}

func (s *SessionStore) UpdateLastActiveByIdentity(ctx context.Context, identity uuid.UUID, t time.Time) error {
	_, err := s.Client(ctx).Session.Update().
		Where(
			session.IdentityEQ(identity),
			session.RevokedAtIsNil(),
			session.ExpiresAtGT(time.Now()),
		).
		SetLastActiveAt(t).
		Save(ctx)

	if err != nil {
		if db.IsNotFound(err) {
			return biz.ErrNotFound
		}
		return err
	}

	return nil
}

func (s *SessionStore) RevokedByID(ctx context.Context, id int) error {
	err := s.Client(ctx).Session.UpdateOneID(id).
		SetRevokedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return biz.ErrNotFound
		}
		return err
	}

	return nil
}

func (s *SessionStore) RevokedByIdentity(ctx context.Context, identity uuid.UUID) (int, error) {
	info, err := s.Client(ctx).Session.Query().
		Where(session.IdentityEQ(identity)).
		Select(session.FieldID).
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return 0, biz.ErrNotFound
		}
		return 0, err
	}

	err = s.Client(ctx).Session.UpdateOneID(info.ID).
		SetRevokedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	return info.ID, nil
}

func (s *SessionStore) RevokeByAppID(ctx context.Context, appID int) ([]int, []uuid.UUID, error) {
	client := s.Client(ctx)

	// 所有未被吊销的会话
	sessions, err := client.Session.Query().
		Where(
			session.AppIDEQ(appID),
			session.RevokedAtIsNil(),
		).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	if len(sessions) == 0 {
		return []int{}, []uuid.UUID{}, nil
	}

	sessionIDs := make([]int, len(sessions))
	identities := make([]uuid.UUID, len(sessions))
	for i, v := range sessions {
		sessionIDs[i] = v.ID
		identities[i] = v.Identity
	}

	// 更新状态为已吊销
	_, err = client.Session.Update().
		Where(session.IDIn(sessionIDs...)).
		SetRevokedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, nil, err
	}

	return sessionIDs, identities, nil
}

func (s *SessionStore) RevokeByUserID(ctx context.Context, userID int) ([]int, []uuid.UUID, error) {
	client := s.Client(ctx)

	sessions, err := client.Session.Query().
		Where(
			session.UserIDEQ(userID),
			session.RevokedAtIsNil(),
		).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	if len(sessions) == 0 {
		return []int{}, []uuid.UUID{}, nil
	}

	sessionIDs := make([]int, len(sessions))
	identities := make([]uuid.UUID, len(sessions))
	for i, v := range sessions {
		sessionIDs[i] = v.ID
		identities[i] = v.Identity
	}

	_, err = client.Session.Update().
		Where(session.IDIn(sessionIDs...)).
		SetRevokedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, nil, err
	}

	return sessionIDs, identities, nil
}

func (s *SessionStore) ClearExpireOrRevoked(ctx context.Context) (int, error) {
	now := time.Now()
	batchSize := 1000

	// 查询满足过期或已撤销条件的会话
	ids, err := s.Client(ctx).Session.Query().
		Where(
			session.Or(
				session.ExpiresAtLTE(now),
				session.RevokedAtNotNil(),
			),
		).
		Limit(batchSize).
		Select(session.FieldID).
		Ints(ctx)

	if err != nil {
		return 0, err
	}

	if len(ids) == 0 {
		return 0, nil
	}

	// 检查当前是否已经在事务中
	_, inTx := txFromContext(ctx)
	client := s.Client(ctx)

	if inTx {
		// 如果已经在事务中，直接复用当前事务客户端执行清理
		return s.executeClearBatch(ctx, client, ids)
	}

	// 如果不在事务中，为这批次清理独立开启一个事务
	tx, err := client.Tx(ctx)
	if err != nil {
		return 0, fmt.Errorf("开启会话清理事务失败: %w", err)
	}

	// 将事务注入 Context
	txCtx := context.WithValue(ctx, txKey{}, tx)
	committed := false

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if !committed {
			_ = tx.Rollback()
		}
	}()

	// 执行删除
	affected, err := s.executeClearBatch(txCtx, tx.Client(), ids)
	if err != nil {
		return 0, err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("提交会话清理事务失败: %w", err)
	}
	committed = true

	return affected, nil
}

func (s *SessionStore) executeClearBatch(ctx context.Context, dbClient *db.Client, sessionIDs []int) (int, error) {
	// 先删除关联的令牌
	_, err := dbClient.RefreshToken.Delete().
		Where(refreshtoken.SessionIDIn(sessionIDs...)).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("清理会话关联的RefreshToken失败: %w", err)
	}

	// 再删除会话本身
	affected, err := dbClient.Session.Delete().
		Where(session.IDIn(sessionIDs...)).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("清理Session失败: %w", err)
	}

	return affected, nil
}

func (s *SessionStore) ClearExpireOrRevokedAll(ctx context.Context) (int, error) {
	totalAffected := 0
	for {
		affected, err := s.ClearExpireOrRevoked(ctx)
		if err != nil {
			return totalAffected, err
		}
		if affected == 0 {
			break // 没有需要清理的数据了，退出循环
		}
		totalAffected += affected

		time.Sleep(10 * time.Millisecond)
	}
	return totalAffected, nil
}
