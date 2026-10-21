package biz

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"budeku/iam/internal/enum"
)

type BizKeyProvider struct {
	keypairStore KeypairStore
}

func NewBizKeyProvider(keypairStore KeypairStore) *BizKeyProvider {
	return &BizKeyProvider{
		keypairStore: keypairStore,
	}
}

func (p *BizKeyProvider) SigningKey(ctx context.Context, appID int) (string, crypto.Signer, error) {
	kp, err := p.keypairStore.GetActiveByAppID(ctx, appID)
	if err != nil {
		return "", nil, err
	}

	block, _ := pem.Decode([]byte(kp.PrivateKey))
	if block == nil {
		return "", nil, fmt.Errorf("解析私钥 PEM 失败: kid=%s", kp.Kid)
	}

	var signer crypto.Signer
	switch kp.Algorithm {
	case enum.KeypairAlgoRS256:
		priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			parsed, err8 := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err8 != nil {
				return "", nil, fmt.Errorf("解析 RSA 私钥失败: %w", err)
			}
			rsaKey, ok := parsed.(*rsa.PrivateKey)
			if !ok {
				return "", nil, errors.New("PKCS8 密钥不是 RSA 类型")
			}
			signer = rsaKey
		} else {
			signer = priv
		}
	case enum.KeypairAlgoES256:
		priv, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			parsed, err8 := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err8 != nil {
				return "", nil, fmt.Errorf("解析 EC 私钥失败: %w", err)
			}
			ecKey, ok := parsed.(*ecdsa.PrivateKey)
			if !ok {
				return "", nil, errors.New("PKCS8 密钥不是 ECDSA 类型")
			}
			signer = ecKey
		} else {
			signer = priv
		}
	default:
		return "", nil, fmt.Errorf("不支持的密钥算法: %d", kp.Algorithm)
	}

	return kp.Kid, signer, nil
}

func (p *BizKeyProvider) VerifyKey(ctx context.Context, kid string) (crypto.PublicKey, error) {
	kp, err := p.keypairStore.GetByKid(ctx, kid)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(kp.PublicKey))
	if block == nil {
		return nil, fmt.Errorf("解析公钥 PEM 失败: kid=%s", kid)
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		if pub1, err1 := x509.ParsePKCS1PublicKey(block.Bytes); err1 == nil {
			return pub1, nil
		}
		return nil, fmt.Errorf("解析公钥失败: %w", err)
	}

	return pub, nil
}
