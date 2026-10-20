package jwt

import (
    "crypto"

    "github.com/lestrrat-go/jwx/v4/jwk"
)

type JWK = jwk.Key
type JWKSet = jwk.Set

func PublicKeyToJWK(kid string, pub crypto.PublicKey) (jwk.Key, error) {
    key, err := jwk.Import[jwk.Key](pub)
    if err != nil {
        return nil, err // 公钥导入失败
    }
    if err := key.Set(jwk.KeyIDKey, kid); err != nil {
        return nil, err // 设置 Key ID 失败
    }
    if err := key.Set(jwk.KeyUsageKey, "sig"); err != nil {
        return nil, err // 设置 Key 用途失败
    }

    return key, nil
}