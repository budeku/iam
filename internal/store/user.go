package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/infra/ent/db/user"
	"context"
	"time"

	"github.com/google/uuid"
)

var _ biz.UserStore = (*UserStore)(nil)

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

func userGetFields() []string {
	return []string{
		user.FieldID,
		user.FieldUsername,
		user.FieldUUID,
		user.FieldEmail,
		user.FieldPassword,
		user.FieldStatus,
		user.FieldLastLoginAt,
		user.FieldIsSuperAdmin,
	}
}

func (s *UserStore) Get(ctx context.Context, id int) (*biz.UserGetResp, error) {
	info, err := s.Client(ctx).User.Query().
		Where(
			user.IDEQ(id),
			user.DeletedAtIsNil(),
		).
		Select(userGetFields()...).
		Only(ctx)

	if err != nil {
		if db.IsNotFound(err) {
			return nil, biz.ErrNotFound
		}
		return nil, err
	}

	return toUserGetResp(info), nil
}

func (s *UserStore) GetByUsername(ctx context.Context, username string) (*biz.UserGetResp, error) {
	info, err := s.Client(ctx).User.Query().
		Where(
			user.UsernameEQ(username),
			user.DeletedAtIsNil(),
		).
		Select(userGetFields()...).
		Only(ctx)

	if err != nil {
		if db.IsNotFound(err) {
			return nil, biz.ErrNotFound
		}
		return nil, err
	}

	return toUserGetResp(info), nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*biz.UserGetResp, error) {
	info, err := s.Client(ctx).User.Query().
		Where(
			user.EmailEQ(email),
			user.DeletedAtIsNil(),
		).
		Select(userGetFields()...).
		Only(ctx)

	if err != nil {
		if db.IsNotFound(err) {
			return nil, biz.ErrNotFound
		}
		return nil, err
	}

	return toUserGetResp(info), nil
}

func (s *UserStore) GetByUUID(ctx context.Context, userUUID uuid.UUID) (*biz.UserGetResp, error) {
	info, err := s.Client(ctx).User.Query().
		Where(
			user.UUIDEQ(userUUID),
			user.DeletedAtIsNil(),
		).
		Select(userGetFields()...).
		Only(ctx)

	if err != nil {
		if db.IsNotFound(err) {
			return nil, biz.ErrNotFound
		}
		return nil, err
	}

	return toUserGetResp(info), nil
}

func (s *UserStore) UpdateLastLogin(ctx context.Context, id int, t time.Time) error {
	err := s.Client(ctx).User.UpdateOneID(id).
		SetLastLoginAt(t).
		Exec(ctx)

	if err != nil {
		if db.IsNotFound(err) {
			return biz.ErrNotFound
		}
		return err
	}

	return nil
}
