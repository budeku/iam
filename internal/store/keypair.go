package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/conf"
	"budeku/iam/internal/enum"
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/infra/ent/db/keypair"
	"context"
	"errors"
	"fmt"
	"time"
)

var _ biz.KeypairStore = (*KeypairStore)(nil)

type KeypairStore struct {
	*baseStore
}

func NewKeypairStore(client *db.Client) *KeypairStore {
	return &KeypairStore{baseStore: newBaseStore(client)}
}

func (s *KeypairStore) GetIAMKeypair(ctx context.Context) (*db.Keypair, error) {
	kp, err := s.Client(ctx).Keypair.Query().
		Where(
			keypair.StatusEQ(enum.KeypairStatusActive),
			keypair.AppIDEQ(conf.IAMAppID),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return kp, nil
}

func (s *KeypairStore) InitIAMKeypair(ctx context.Context, kid string, algorithm enum.KeypairAlgorithm, publicKey, privateKey string) (*db.Keypair, error) {
	kp, err := s.Client(ctx).Keypair.Create().
		SetAppID(conf.IAMAppID).
		SetKid(kid).
		SetAlgorithm(algorithm).
		SetPublicKey(publicKey).
		SetPrivateKey(privateKey).
		SetStatus(enum.KeypairStatusActive).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return kp, nil
}

func keypairListFields() []string {
	return []string{
		keypair.FieldID,
		keypair.FieldAppID,
		keypair.FieldKid,
		keypair.FieldAlgorithm,
		keypair.FieldStatus,
		keypair.FieldActivatedAt,
		keypair.FieldCreatedAt,
	}
}

func (s *KeypairStore) List(ctx context.Context, page, perPage int) ([]biz.KeypairListItemResp, int, error) {
	q := s.Client(ctx).Keypair.Query()

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []biz.KeypairListItemResp{}, 0, nil
	}

	offset, limit := paginate(total, page, perPage)

	list, err := q.
		Order(db.Desc(keypair.FieldID)).
		Select(keypairListFields()...).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return toKeypairList(list), total, nil
}

func (s *KeypairStore) Get(ctx context.Context, id int) (*biz.KeypairGetResp, error) {
	info, err := s.Client(ctx).Keypair.Query().
		Where(keypair.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, biz.ErrNotFound
		}
		return nil, err
	}

	return toKeypairGetResp(info), nil
}

func (s *KeypairStore) GetByAppID(ctx context.Context, appID int) (*biz.KeypairGetResp, error) {
	info, err := s.Client(ctx).Keypair.Query().
		Where(
			keypair.AppID(appID),
			keypair.Status(enum.KeypairStatusActive),
		).
		Order(db.Desc(keypair.FieldActivatedAt)).
		First(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, biz.ErrNotFound
		}
		return nil, err
	}

	return toKeypairGetResp(info), nil
}

func (s *KeypairStore) GetByKid(ctx context.Context, kid string) (*biz.KeypairGetResp, error) {
	info, err := s.Client(ctx).Keypair.Query().
		Where(
			keypair.Kid(kid),
			keypair.StatusIn(enum.KeypairStatusActive, enum.KeypairStatusGrace),
		).
		First(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, biz.ErrNotFound
		}
		return nil, err
	}

	return toKeypairGetResp(info), nil
}

func (s *KeypairStore) GetAllValidByAppID(ctx context.Context, appID int) ([]biz.KeypairGetResp, error) {
	list, err := s.Client(ctx).Keypair.Query().
		Where(
			keypair.AppIDEQ(appID),
			keypair.StatusIn(enum.KeypairStatusActive, enum.KeypairStatusGrace),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return toKeypairGetAllResp(list), nil
}

func (s *KeypairStore) Create(
	ctx context.Context,
	appID int,
	kid string,
	algorithm enum.KeypairAlgorithm,
	status enum.KeypairStatus,
	publicKey, privateKey string,
) (*biz.KeypairCreateResp, error) {
	info, err := s.Client(ctx).Keypair.Create().
		SetAppID(appID).
		SetKid(kid).
		SetAlgorithm(algorithm).
		SetPublicKey(publicKey).
		SetPrivateKey(privateKey).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return toKeypairCreate(info), nil
}

func (s *KeypairStore) Downgrade(ctx context.Context, id int) error {
	info, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	if info.Status != enum.KeypairStatusActive {
		return fmt.Errorf("当前密钥状态为 %s，无法降级", info.Status)
	}

	if err := s.Client(ctx).Keypair.UpdateOneID(id).
		SetStatus(enum.KeypairStatusGrace).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *KeypairStore) Retire(ctx context.Context, id int) error {
	info, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	if info.Status != enum.KeypairStatusActive && info.Status != enum.KeypairStatusGrace {
		return fmt.Errorf("当前密钥状态为 %s，无法执行作废操作", info.Status)
	}

	validKeys, err := s.GetAllValidByAppID(ctx, info.AppID)
	if err != nil {
		return err
	}

	if len(validKeys) <= 1 {
		return errors.New("当前应用仅剩最后一个有效密钥，无法作废，请先创建新密钥后再操作")
	}

	if err := s.Client(ctx).Keypair.UpdateOneID(id).
		SetStatus(enum.KeypairStatusRetired).
		SetRetireAt(time.Now()).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}
