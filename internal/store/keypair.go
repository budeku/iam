package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/conf"
	"budeku/iam/internal/enum"
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/infra/ent/db/keypair"
	"context"
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

func (s *KeypairStore) GetActiveByAppID(ctx context.Context, appID int) (*biz.KeypairGetResp, error) {
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
		Where(keypair.Kid(kid)).
		First(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, biz.ErrNotFound
		}
	}

	return toKeypairGetResp(info), nil
}
