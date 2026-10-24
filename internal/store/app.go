package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/infra/ent/db/app"
	"context"
	"time"

	"github.com/google/uuid"
)

var _ biz.AppStore = (*AppStore)(nil)

type AppStore struct {
	*baseStore
}

func NewAppStore(client *db.Client) *AppStore {
	return &AppStore{newBaseStore(client)}
}

func appListFields() []string {
	return []string{
		app.FieldID,
		app.FieldName,
		app.FieldClientID,
		app.FieldStatus,
		app.FieldCreatedAt,
	}
}

func (s *AppStore) List(ctx context.Context, keyword string, page, perPage int) ([]biz.AppListItemResp, int, error) {
	q := s.Client(ctx).App.Query().
		Where(app.DeletedAtIsNil())

	if keyword != "" {
		q = q.Where(
			app.Or(
				app.NameContains(keyword),
			),
		)
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []biz.AppListItemResp{}, 0, nil
	}

	offset, limit := paginate(total, page, perPage)

	list, err := q.
		Order(db.Desc(app.FieldID)).
		Select(appListFields()...).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return toAppList(list), total, nil
}

func (s *AppStore) ListByIds(ctx context.Context, ids []int) ([]biz.AppListItemResp, error) {
	if len(ids) == 0 {
		return []biz.AppListItemResp{}, nil
	}

	list, err := s.Client(ctx).App.Query().
		Where(app.IDIn(ids...), app.DeletedAtIsNil()).
		Select(appListFields()...).
		Order(db.Desc(app.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return toAppList(list), nil
}

func (s *AppStore) Options(ctx context.Context) ([]biz.AppOptionItemResp, error) {
	list, err := s.Client(ctx).App.Query().
		Where(
			app.DeletedAtIsNil(),
			app.StatusEQ(true),
		).
		Select(app.FieldID, app.FieldName).
		Order(db.Asc(app.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return toAppOptions(list), nil
}

func (s *AppStore) Get(ctx context.Context, id int) (*biz.AppGetResp, error) {
	info, err := s.Client(ctx).App.Query().
		Where(app.IDEQ(id), app.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, biz.ErrNotFound
		}
		return nil, err
	}
	return toAppGet(info), nil
}

func (s *AppStore) GetByClientID(ctx context.Context, clientID uuid.UUID) (*biz.AppGetResp, error) {
	info, err := s.Client(ctx).App.Query().
		Where(app.ClientIDEQ(clientID), app.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, biz.ErrNotFound
		}
		return nil, err
	}
	return toAppGet(info), nil
}

func (s *AppStore) CreateDuplicate(ctx context.Context, name string, clientID uuid.UUID) (bool, error) {
	q := s.Client(ctx).App.Query().
		Where(
			app.DeletedAtIsNil(),
			app.Or(
				app.NameEQ(name),
				app.ClientIDEQ(clientID),
			),
		)

	exist, err := q.Exist(ctx)
	if err != nil {
		return false, err
	}
	return exist, nil
}

func (s *AppStore) UpdateDuplicate(ctx context.Context, id int, name string) (bool, error) {
	q := s.Client(ctx).App.Query().
		Where(
			app.DeletedAtIsNil(),
			app.NameEQ(name),
			app.IDNEQ(id),
		)

	exist, err := q.Exist(ctx)
	if err != nil {
		return false, err
	}
	return exist, nil
}

func (s *AppStore) Create(ctx context.Context, name string, clientID uuid.UUID, redirectUris []string, status bool) (*biz.AppGetResp, error) {
	info, err := s.Client(ctx).App.Create().
		SetName(name).
		SetClientID(clientID).
		SetRedirectUris(redirectUris).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return toAppGet(info), nil
}

func (s *AppStore) Update(ctx context.Context, id int, name string, redirectUris []string, status bool) (*biz.AppGetResp, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return nil, err
	}

	info, err := s.Client(ctx).App.UpdateOneID(id).
		SetName(name).
		SetRedirectUris(redirectUris).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return toAppGet(info), nil
}

func (s *AppStore) UpdateStatus(ctx context.Context, id int, status bool) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}

	err := s.Client(ctx).App.UpdateOneID(id).
		SetStatus(status).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *AppStore) Delete(ctx context.Context, id int) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}

	if err := s.Client(ctx).App.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
