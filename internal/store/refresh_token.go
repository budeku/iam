package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/infra/ent/db"
	"context"
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
		SetIdentity(body.Identify).
		SetJti(body.Jti).
		SetExpiresAt(body.ExpiresAt).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return toRefreshTokenCreateResp(info), nil
}
