package biz

import (
	"budeku/iam/internal/enum"
	"context"
	"time"
)

type KeypairGetResp struct {
	ID          int
	AppID       int
	Kid         string
	Algorithm   enum.KeypairAlgorithm
	PublicKey   string
	PrivateKey  string
	ActivatedAt time.Time
	Status      enum.KeypairStatus
	RetireAt    *time.Time
}

type KeypairStore interface {
	GetActiveByAppID(ctx context.Context, appID int) (*KeypairGetResp, error)
	GetByKid(ctx context.Context, kid string) (*KeypairGetResp, error)
}

type KeypairBiz struct {
	store KeypairStore
}

func NewKeypairBiz(store KeypairStore) *KeypairBiz {
	return &KeypairBiz{store: store}
}
