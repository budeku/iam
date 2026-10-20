package store

import (
	"budeku/iam/internal/enum"
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/infra/ent/db/keypair"
	"context"
)

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
			keypair.AppIDEQ(0),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return kp, nil
}

func (s *KeypairStore) InitIAMKeypair(ctx context.Context, kid string, algorithm enum.KeypairAlgorithm, publicKey, privateKey string) (*db.Keypair, error) {
	kp, err := s.Client(ctx).Keypair.Create().
		SetAppID(0).
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
