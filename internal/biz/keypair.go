package biz

import (
	"budeku/iam/internal/conf"
	"budeku/iam/internal/enum"
	"budeku/iam/internal/infra/redis"
	"budeku/iam/internal/jwt"
	"budeku/iam/pkg/cryptoutil"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v4/jwk"
)

type KeypairGetResp struct {
	ID          int                   `json:"id"`
	AppID       int                   `json:"app_id"`
	Kid         uuid.UUID             `json:"kid"`
	Algorithm   enum.KeypairAlgorithm `json:"algorithm"`
	PublicKey   string                `json:"public_key"`
	PrivateKey  string                `json:"private_key"`
	Status      enum.KeypairStatus    `json:"status"`
	ActivatedAt time.Time             `json:"activated_at"`
	RetireAt    *time.Time            `json:"retire_at"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

type KeypairListItemResp struct {
	ID          int                   `json:"id"`
	AppID       int                   `json:"app_id"`
	Kid         uuid.UUID             `json:"kid"`
	Algorithm   enum.KeypairAlgorithm `json:"algorithm"`
	Status      enum.KeypairStatus    `json:"status"`
	ActivatedAt time.Time             `json:"activated_at"`
	CreatedAt   time.Time             `json:"created_at"`
}

type KeypairCreateResp struct {
	ID int `json:"id"`
}

type KeypairStore interface {
	List(ctx context.Context, page, perPage int) ([]KeypairListItemResp, int, error)

	Get(ctx context.Context, id int) (*KeypairGetResp, error)
	GetByAppID(ctx context.Context, appID int) (*KeypairGetResp, error)
	GetByKid(ctx context.Context, kid uuid.UUID) (*KeypairGetResp, error)
	GetAllValidByAppID(ctx context.Context, appID int) ([]KeypairGetResp, error)

	Create(ctx context.Context, appID int, kid uuid.UUID, algorithm enum.KeypairAlgorithm, status enum.KeypairStatus, publicKey, privateKey string) (*KeypairCreateResp, error)

	Downgrade(ctx context.Context, id int) error
	Retire(ctx context.Context, id int) error
}

type KeypairBiz struct {
	store      KeypairStore
	appStore   AppStore
	transactor Transactor
	rdb        *redis.Redis
}

func NewKeypairBiz(
	store KeypairStore,
	appStore AppStore,
	transactor Transactor,
) *KeypairBiz {
	return &KeypairBiz{
		store:      store,
		appStore:   appStore,
		transactor: transactor,
		rdb:        redis.DB(),
	}
}

func jwksCacheKey(appID int) string {
	return fmt.Sprintf("keypair:jwks:%d", appID)
}

type cachedPublicKeyItem struct {
	Kid       uuid.UUID          `json:"kid"`
	PublicKey string             `json:"public_key"`
	Status    enum.KeypairStatus `json:"status"`
}

type jwksCacheWrapper struct {
	IsEmpty bool                  `json:"is_empty"`
	Items   []cachedPublicKeyItem `json:"items"`
}

func (b *KeypairBiz) JWKS(ctx context.Context, clientID uuid.UUID) (jwt.JWKSet, error) {
	appInfo, err := b.appStore.GetByClientID(ctx, clientID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return jwk.NewSet(), nil
		}
		return nil, err
	}

	cacheKey := jwksCacheKey(appInfo.ID)

	var cachedWrapper jwksCacheWrapper
	err = b.rdb.GetJSON(ctx, cacheKey, &cachedWrapper)
	if err == nil {
		if cachedWrapper.IsEmpty {
			return jwk.NewSet(), nil
		}
		if len(cachedWrapper.Items) > 0 {
			return b.buildJWKSFromCache(cachedWrapper.Items)
		}
	}

	// 应用存在
	dbKeypairs, err := b.store.GetAllValidByAppID(ctx, appInfo.ID)
	if err != nil {
		return nil, err
	}

	// 没有 keypair
	if len(dbKeypairs) == 0 {
		emptyWrapper := jwksCacheWrapper{IsEmpty: true, Items: []cachedPublicKeyItem{}}
		_ = b.rdb.SetJSON(ctx, cacheKey, emptyWrapper, 30*time.Second)
		return jwk.NewSet(), nil
	}

	// 缓存 keypair
	cachedPubKeys := make([]cachedPublicKeyItem, 0, len(dbKeypairs))
	for _, kp := range dbKeypairs {
		cachedPubKeys = append(cachedPubKeys, cachedPublicKeyItem{
			Kid:       kp.Kid,
			PublicKey: kp.PublicKey,
			Status:    kp.Status,
		})
	}

	validWrapper := jwksCacheWrapper{IsEmpty: false, Items: cachedPubKeys}
	_ = b.rdb.SetJSON(ctx, cacheKey, validWrapper, 15*time.Minute)

	return b.buildJWKSFromCache(cachedPubKeys)
}

func (b *KeypairBiz) buildJWKSFromCache(items []cachedPublicKeyItem) (jwt.JWKSet, error) {
	keySet := jwk.NewSet()
	for _, item := range items {
		pubKey, err := parsePublicKeyPEM(item.PublicKey)
		if err != nil {
			continue
		}

		jwkKey, err := jwt.PublicKeyToJWK(item.Kid.String(), pubKey)
		if err != nil {
			continue
		}

		if err := keySet.AddKey(jwkKey); err != nil {
			continue
		}
	}
	return keySet, nil
}

func (b *KeypairBiz) List(ctx context.Context, page, perPage int) ([]KeypairListItemResp, int, error) {
	return b.store.List(ctx, page, perPage)
}

func (b *KeypairBiz) Info(ctx context.Context, id int) (*KeypairGetResp, error) {
	resp, err := b.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if resp.PrivateKey != "" {
		decrypted, err := decryptPrivateKey(resp.PrivateKey)
		if err == nil {
			resp.PrivateKey = decrypted
		}
	}

	return resp, nil
}

func decryptPrivateKey(encryptedBase64 string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", err
	}
	plaintextBytes, err := cryptoutil.Decrypt(ciphertext, []byte(conf.Get().Security.MasterKey))
	if err != nil {
		return "", err
	}
	return string(plaintextBytes), nil
}

func (b *KeypairBiz) evictKeypairCaches(ctx context.Context, appID int, kid uuid.UUID) {
	_ = b.rdb.Del(ctx, jwksCacheKey(appID))
	_ = b.rdb.Del(ctx, fmt.Sprintf("keypair:active:%d", appID))

	if kid != uuid.Nil {
		_ = b.rdb.Del(ctx, fmt.Sprintf("keypair:pub:%s", kid.String()))
	}
}

func (b *KeypairBiz) Create(ctx context.Context, appID int, algorithm enum.KeypairAlgorithm, status enum.KeypairStatus) (*KeypairCreateResp, error) {
	appInfo, err := b.appStore.Get(ctx, appID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, err
		}
		return nil, err
	}
	if !appInfo.Status {
		return nil, errors.New("应用已被禁用")
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// 生成公钥 PEM
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, err
	}
	pubPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDER,
	}))

	// 生成私钥 PEM
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, err
	}
	privPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privDER,
	}))

	encryptedBytes, err := cryptoutil.Encrypt([]byte(privPEM), []byte(conf.Get().Security.MasterKey))
	if err != nil {
		return nil, err
	}
	encryptedPrivBase64 := base64.StdEncoding.EncodeToString(encryptedBytes)

	var createResp *KeypairCreateResp
	// 随机生成 kid
	kid := uuid.New()

	err = b.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if status == enum.KeypairStatusActive {
			existingActive, getErr := b.store.GetByAppID(txCtx, appID)
			if getErr == nil && existingActive != nil {
				if downgradeErr := b.store.Downgrade(txCtx, existingActive.ID); downgradeErr != nil {
					return downgradeErr
				}
			} else if getErr != nil && !errors.Is(getErr, ErrNotFound) {
				return getErr
			}
		}

		resp, createErr := b.store.Create(txCtx, appID, kid, algorithm, status, pubPEM, encryptedPrivBase64)
		if createErr != nil {
			return createErr
		}

		createResp = resp
		return nil
	})

	if err != nil {
		return nil, err
	}

	b.evictKeypairCaches(ctx, appID, kid)

	return createResp, nil
}

func (b *KeypairBiz) Downgrade(ctx context.Context, id int) error {
	info, err := b.store.Get(ctx, id)
	if err != nil {
		return err
	}

	if err := b.store.Downgrade(ctx, id); err != nil {
		return err
	}

	b.evictKeypairCaches(ctx, info.AppID, info.Kid)
	return nil
}

func (b *KeypairBiz) Retire(ctx context.Context, id int) error {
	info, err := b.store.Get(ctx, id)
	if err != nil {
		return err
	}

	if err := b.store.Retire(ctx, id); err != nil {
		return err
	}

	b.evictKeypairCaches(ctx, info.AppID, info.Kid)

	return nil
}
