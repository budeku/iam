package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/infra/ent/db/refreshtoken"
	"budeku/iam/internal/infra/ent/db/session"
	"budeku/iam/internal/infra/ent/db/user"
	"context"
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

	return toSessionGetResp(info), nil
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

func (s *SessionStore) Create(ctx context.Context, body biz.SessionCommand) (*biz.SessionCreateResp, error) {
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

	return toSessionCreateResp(info), nil
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
