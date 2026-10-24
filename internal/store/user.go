package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/infra/ent/db/refreshtoken"
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

func userListFields() []string {
	return []string{
		user.FieldID,
		user.FieldUsername,
		user.FieldEmail,
		user.FieldUUID,
		user.FieldStatus,
		user.FieldLastLoginAt,
		user.FieldIsSuperAdmin,
	}
}

func (s *UserStore) List(ctx context.Context, keyword string, page, perPage int) ([]biz.UserListItemResp, int, error) {
	q := s.Client(ctx).User.Query().
		Where(user.DeletedAtIsNil())

	if keyword != "" {
		q = q.Where(
			user.Or(
				user.UsernameContains(keyword),
				user.EmailContains(keyword),
			),
		)
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []biz.UserListItemResp{}, 0, nil
	}

	offset, limit := paginate(total, page, perPage)

	list, err := q.
		Order(db.Desc(user.FieldID)).
		Select(userListFields()...).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return toUserList(list), total, nil
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

	return toUserGet(info), nil
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

	return toUserGet(info), nil
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

	return toUserGet(info), nil
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

	return toUserGet(info), nil
}

func (s *UserStore) GetUsersWithActiveTokens(ctx context.Context) ([]*db.User, error) {
	return s.Client(ctx).User.Query().
		Where(
			user.HasRefreshTokensWith(
				refreshtoken.RevokedAtIsNil(),
				refreshtoken.ExpiresAtGT(time.Now()),
			),
		).
		All(ctx)
}

func (s *UserStore) Duplicate(ctx context.Context, username, email string, id ...int) (bool, error) {
	q := s.Client(ctx).User.Query().
		Where(
			user.DeletedAtIsNil(),
			user.Or(
				user.UsernameEQ(username),
				user.EmailEQ(email),
			),
		)
	if len(id) > 0 && id[0] > 0 {
		q = q.Where(user.IDNEQ(id[0]))
	}

	exist, err := q.Exist(ctx)
	if err != nil {
		return false, err
	}

	return exist, nil
}

func (s *UserStore) Create(ctx context.Context, username, email, passwordHash string, uuid uuid.UUID, status bool) (*biz.UserCreateResp, error) {
	info, err := s.Client(ctx).User.Create().
		SetUsername(username).
		SetEmail(email).
		SetPassword([]byte(passwordHash)).
		SetUUID(uuid).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return toUserCreate(info), nil
}

func (s *UserStore) Update(ctx context.Context, id int, username, email, passwordHash string, status bool) (*biz.UserGetResp, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return nil, err
	}

	query := s.Client(ctx).User.UpdateOneID(id).
		Where(user.IsSuperAdminEQ(false)).
		SetUsername(username).
		SetEmail(email).
		SetStatus(status)

	if passwordHash != "" {
		query.SetPassword([]byte(passwordHash))
	}

	info, err := query.Save(ctx)
	if err != nil {
		return nil, err
	}

	return toUserGet(info), nil
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

func (s *UserStore) UpdateStatus(ctx context.Context, id int, status bool) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}

	err := s.Client(ctx).User.UpdateOneID(id).
		Where(user.IsSuperAdminEQ(false)).
		SetStatus(status).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) Delete(ctx context.Context, id int) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}

	if err := s.Client(ctx).User.UpdateOneID(id).
		Where(user.IsSuperAdminEQ(false)).
		SetDeletedAt(time.Now()).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
