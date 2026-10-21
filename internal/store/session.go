package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/infra/ent/db"
	"context"
	"time"
)

var _ biz.SessionStore = (*SessionStore)(nil)

type SessionStore struct {
	*baseStore
}

func NewSessionStore(client *db.Client) *SessionStore {
	return &SessionStore{newBaseStore(client)}
}

func (s *SessionStore) Create(ctx context.Context, body biz.SessionCommand) (*biz.SessionCreateResp, error) {
	info, err := s.Client(ctx).Session.Create().
		SetIdentity(body.Identify).
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
