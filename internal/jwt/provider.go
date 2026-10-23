package jwt

import (
	"context"
	"crypto"
	"errors"

	"github.com/google/uuid"
)

var ErrKeyNotFound = errors.New("jwt 键不存在")

type KeyProvider interface {
	SigningKey(ctx context.Context, appID int) (kid uuid.UUID, priv crypto.Signer, err error)
	VerifyKey(ctx context.Context, kid uuid.UUID) (pub crypto.PublicKey, err error)
}

type MemoryKeyProvider struct {
	appID int
	kid   uuid.UUID
	priv  crypto.Signer
}

func NewMemoryKeyProvider(appID int, kid uuid.UUID, priv crypto.Signer) *MemoryKeyProvider {
	return &MemoryKeyProvider{appID: appID, kid: kid, priv: priv}
}

func (p *MemoryKeyProvider) SigningKey(_ context.Context, appID int) (uuid.UUID, crypto.Signer, error) {
	if appID != p.appID {
		return uuid.Nil, nil, ErrKeyNotFound
	}
	return p.kid, p.priv, nil
}

func (p *MemoryKeyProvider) VerifyKey(_ context.Context, kid uuid.UUID) (crypto.PublicKey, error) {
	if kid != p.kid {
		return nil, ErrKeyNotFound // Key ID 不匹配
	}
	return p.priv.Public(), nil
}
