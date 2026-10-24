package biz

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"budeku/iam/internal/conf"
	"budeku/iam/internal/consts"
	"budeku/iam/internal/enum"
	"budeku/iam/internal/infra/redis"
	"budeku/iam/pkg/cryptoutil"

	"github.com/google/uuid"
)

type BizKeyProvider struct {
	keypairStore KeypairStore
	rdb          *redis.Redis
}

type cachedKeypair struct {
	Kid        uuid.UUID             `json:"kid"`
	PrivateKey string                `json:"private_key"`
	Algorithm  enum.KeypairAlgorithm `json:"algorithm"`
}

func NewBizKeyProvider(keypairStore KeypairStore) *BizKeyProvider {
	return &BizKeyProvider{
		keypairStore: keypairStore,
		rdb:          redis.DB(),
	}
}

func activeKeypairCacheKey(appID int) string {
	return fmt.Sprintf("%s:%d", consts.RedisKeypair, appID)
}

func publicCertCacheKey(kid uuid.UUID) string {
	return fmt.Sprintf("%s:%s", consts.RedisKeypairPublicKey, kid.String())
}

// 私钥签名（Signing）
func (p *BizKeyProvider) SigningKey(ctx context.Context, appID int) (uuid.UUID, crypto.Signer, error) {
	cacheKey := activeKeypairCacheKey(appID)

	var kp cachedKeypair

	err := p.rdb.GetJSON(ctx, cacheKey, &kp)
	if err == nil && kp.Kid != uuid.Nil {
		signer, parseErr := parsePrivateKey(kp.PrivateKey, kp.Algorithm)
		if parseErr == nil {
			return kp.Kid, signer, nil
		}
	}

	// 缓存未命中，回源查数据库
	var dbKp *KeypairGetResp

	// TODO: 暂时这样区分下，目前没有好的想法
	if appID == conf.IAMAppID {
		dbKp, err = p.keypairStore.GetIAMKeypair(ctx)
	} else {
		dbKp, err = p.keypairStore.GetByAppID(ctx, appID)
	}

	if err != nil {
		return uuid.Nil, nil, err
	}

	// 从数据库读取后进行解密
	ciphertext, err := base64.StdEncoding.DecodeString(dbKp.PrivateKey)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("私钥 Base64 解码失败: %w", err)
	}

	plaintextPEMBytes, err := cryptoutil.Decrypt(ciphertext, []byte(conf.Get().Security.MasterKey))
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("私钥解密失败: %w", err)
	}
	privPEM := string(plaintextPEMBytes)

	// 解析私钥
	signer, err := parsePrivateKey(privPEM, dbKp.Algorithm)
	if err != nil {
		return uuid.Nil, nil, err
	}

	// 写入缓存
	cacheData := cachedKeypair{
		Kid:        dbKp.Kid,
		PrivateKey: privPEM,
		Algorithm:  dbKp.Algorithm,
	}
	_ = p.rdb.SetJSON(ctx, cacheKey, cacheData, 5*time.Minute)

	return dbKp.Kid, signer, nil
}

// 公钥验签（Verifying）
func (p *BizKeyProvider) VerifyKey(ctx context.Context, kid uuid.UUID) (crypto.PublicKey, error) {
	cacheKey := publicCertCacheKey(kid)

	var pubPEM string
	pubPEM, err := p.rdb.Get(ctx, cacheKey)
	if err == nil && pubPEM != "" {
		// 防缓存穿透
		if pubPEM == "__NULL__" {
			return nil, ErrNotFound
		}

		pub, parseErr := parsePublicKeyPEM(pubPEM)
		if parseErr == nil {
			return pub, nil
		}

		_ = p.rdb.Del(ctx, cacheKey)
	}

	// 缓存未命中，回源查数据库
	dbKp, err := p.keypairStore.GetByKid(ctx, kid)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			_ = p.rdb.Set(ctx, cacheKey, "__NULL__", 2*time.Minute)
		}
		return nil, err
	}

	// 解析公钥
	pub, err := parsePublicKeyPEM(dbKp.PublicKey)
	if err != nil {
		_ = p.rdb.Del(ctx, cacheKey)
		return nil, err
	}

	// 缓存
	_ = p.rdb.Set(ctx, cacheKey, dbKp.PublicKey, 24*time.Hour)

	return pub, nil
}

func parsePrivateKey(privPEM string, algo enum.KeypairAlgorithm) (crypto.Signer, error) {
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, fmt.Errorf("解析私钥 PEM 失败")
	}

	var signer crypto.Signer
	switch algo {
	case enum.KeypairAlgoRS256:
		priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			parsed, err8 := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err8 != nil {
				return nil, fmt.Errorf("解析 RSA 私钥失败: %w", err)
			}
			rsaKey, ok := parsed.(*rsa.PrivateKey)
			if !ok {
				return nil, errors.New("PKCS8 密钥不是 RSA 类型")
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
				return nil, fmt.Errorf("解析 EC 私钥失败: %w", err)
			}
			ecKey, ok := parsed.(*ecdsa.PrivateKey)
			if !ok {
				return nil, errors.New("PKCS8 密钥不是 ECDSA 类型")
			}
			signer = ecKey
		} else {
			signer = priv
		}
	default:
		return nil, fmt.Errorf("不支持的密钥算法: %d", algo)
	}

	return signer, nil
}

func parsePublicKeyPEM(pubPEM string) (crypto.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, fmt.Errorf("解析公钥 PEM 失败")
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
