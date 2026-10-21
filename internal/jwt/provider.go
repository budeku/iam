package jwt

import (
	"context"
	"crypto"
	"errors"
)

var ErrKeyNotFound = errors.New("jwt 键不存在")

type KeyProvider interface {
	SigningKey(ctx context.Context, appID int) (kid string, priv crypto.Signer, err error)
	VerifyKey(ctx context.Context, kid string) (pub crypto.PublicKey, err error)
}

type MemoryKeyProvider struct {
	kid  string
	priv crypto.Signer
}

func NewMemoryKeyProvider(kid string, priv crypto.Signer) *MemoryKeyProvider {
	return &MemoryKeyProvider{kid: kid, priv: priv}
}

func (p *MemoryKeyProvider) SigningKey(_ context.Context, appID int) (string, crypto.Signer, error) {
	return p.kid, p.priv, nil
}

func (p *MemoryKeyProvider) VerifyKey(_ context.Context, kid string) (crypto.PublicKey, error) {
	if kid != p.kid {
		return nil, ErrKeyNotFound // Key ID 不匹配
	}
	return p.priv.Public(), nil
}
