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

func (s *RefreshTokenStore) GetByIdentity(ctx context.Context, identity []byte) (*biz.RefreshTokenGetResp, error) {
	info, err := s.Client(ctx).RefreshToken.Query().
		Where(refreshtoken.IdentityEQ(identity)).
		Only(ctx)

	if err != nil {
		if db.IsNotFound(err) {
			return nil, biz.ErrNotFound
		}
		return nil, err
	}

	return toRefreshTokenGetResp(info), nil
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

func (s *RefreshTokenStore) RevokeByID(ctx context.Context, id int) error {
	err := s.Client(ctx).RefreshToken.UpdateOneID(id).
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

func (s *RefreshTokenStore) RevokedBySessionID(ctx context.Context, sessionID int) error {
	_, err := s.Client(ctx).RefreshToken.Update().
		Where(refreshtoken.SessionIDEQ(sessionID)).
		SetRevokedAt(time.Now()).
		Save(ctx)
	return err
}

func (s *RefreshTokenStore) RevokedBySessionIDs(ctx context.Context, sessionIDs []int) error {
	if len(sessionIDs) == 0 {
		return nil
	}

	_, err := s.Client(ctx).RefreshToken.Update().
		Where(refreshtoken.SessionIDIn(sessionIDs...)).
		SetRevokedAt(time.Now()).
		Save(ctx)

	return err
}
