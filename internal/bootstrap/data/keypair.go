package data

import (
	"budeku/iam/internal/conf"
	"budeku/iam/internal/enum"
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/store"
	"budeku/iam/pkg/cryptoutil"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"github.com/google/uuid"
)

func initKeypair(ctx context.Context, kpStore *store.KeypairStore, result *Result) error {
	if _, err := kpStore.GetIAMKeypair(ctx); err == nil {
		return nil
	} else if !db.IsNotFound(err) {
		return err
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// 生成公钥 PEM (PKIX 格式)
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return err
	}
	pubPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDER,
	}))

	// 生成私钥 PEM (PKCS#8 格式)
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}
	privPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privDER,
	}))

	encryptedBytes, err := cryptoutil.Encrypt([]byte(privPEM), []byte(conf.Get().Security.MasterKey))
	if err != nil {
		return err
	}
	encryptedPrivBase64 := base64.StdEncoding.EncodeToString(encryptedBytes)

	kid := uuid.New()
	if _, err := kpStore.InitIAMKeypair(ctx, kid, enum.KeypairAlgoRS256, pubPEM, encryptedPrivBase64); err != nil {
		return err
	}

	result.KeypairCreated = true
	result.KeypairKid = kid.String()
	return nil
}
