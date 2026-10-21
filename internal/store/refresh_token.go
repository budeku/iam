package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/infra/ent/db/refreshtoken"
	"context"
	"time"
)

var _ biz.RefreshTokenStore = (*RefreshTokenStore)(nil)

type RefreshTokenStore struct {
	*baseStore
}

func NewRefreshTokenStore(client *db.Client) *RefreshTokenStore {
	return &RefreshTokenStore{newBaseStore(client)}
}

func (s *RefreshTokenStore) Create(ctx context.Context, body biz.RefreshTokenCommand) (*biz.RefreshTokenCreateResp, error) {
	info, err := s.Client(ctx).RefreshToken.Create().
		SetAppID(body.AppID).
		SetSessionID(body.SessionID).
		SetUserID(body.UserID).
		SetIdentity(body.Identity).
		SetJti(body.Jti).
		SetExpiresAt(body.ExpiresAt).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return toRefreshTokenCreateResp(info), nil
}

func (s *RefreshTokenStore) RevokedBySessionID(ctx context.Context, sessionID int) error {
	_, err := s.Client(ctx).RefreshToken.Update().
		Where(refreshtoken.SessionIDEQ(sessionID)).
		SetRevokedAt(time.Now()).
		Save(ctx)
	return err
}
