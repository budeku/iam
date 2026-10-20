package store

import (
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/infra/ent/db/user"
	"context"
)

type UserStore struct {
	*baseStore
}

func NewUserStore(client *db.Client) *UserStore {
	return &UserStore{baseStore: newBaseStore(client)}
}

func (s *UserStore) GetSuperAdmin(ctx context.Context) (*db.User, error) {
	info, err := s.Client(ctx).User.Query().
		Where(
			user.IsSuperAdminEQ(true),
			user.DeletedAtIsNil(),
		).Only(ctx)

	if err != nil {
		return nil, err
	}
	return info, nil
}

func (s *UserStore) InitSuperAdmin(ctx context.Context, username, email, passwordHash string) (*db.User, error) {
	client := s.Client(ctx)

	existing, err := client.User.Query().
		Where(user.IsSuperAdminEQ(true)).
		ForUpdate().
		Only(ctx)
	if err == nil {
		return existing, nil
	}
	if !db.IsNotFound(err) {
		return nil, err
	}

	res, err := client.User.Create().
		SetUsername(username).
		SetEmail(email).
		SetPassword([]byte(passwordHash)).
		SetStatus(true).
		SetIsSuperAdmin(true).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}
