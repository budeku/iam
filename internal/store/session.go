package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/infra/ent/db/session"
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

func (s *SessionStore) GetByID(ctx context.Context, id int) (*biz.SessionInfoResp, error) {
	info, err := s.Client(ctx).Session.Query().
		Where(session.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, biz.ErrNotFound
		}
		return nil, err
	}

	return toSessionInfoResp(info), nil
}

func (s *SessionStore) Create(ctx context.Context, body biz.SessionCommand) (*biz.SessionCreateResp, error) {
	info, err := s.Client(ctx).Session.Create().
		SetIdentity(body.Identity).
		SetAppID(body.AppID).
		SetUserID(body.UserID).
		SetIP(body.IP).
		SetUserAgent(body.UserAgent).
		SetAuthMethod(body.AuthMedhod).
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

func (s *SessionStore) RevokedByIdentity(ctx context.Context, identity uuid.UUID) (*biz.SessionInfoResp, error) {
	info, err := s.Client(ctx).Session.Query().
		Where(session.IdentityEQ(identity)).
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, biz.ErrNotFound
		}
		return nil, err
	}

	err = s.Client(ctx).Session.UpdateOneID(info.ID).
		SetRevokedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	return toSessionInfoResp(info), nil
}
