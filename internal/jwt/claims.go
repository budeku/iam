package jwt

import (
	"time"

	"github.com/google/uuid"
	jwtlib "github.com/lestrrat-go/jwx/v4/jwt"
)

type Claims struct {
	jwtlib.Token

	UserUUID  uuid.UUID
	SessionID uuid.UUID
	Name      string

	Fingerprint string // 客户端环境指纹
}

func (c *Claims) ExpiresAt() time.Time {
	exp, _ := c.Token.Expiration()
	return exp
}

func ParseClaims(token jwtlib.Token) (*Claims, error) {
	uidStr, err := jwtlib.Get[string](token, "uid")
	if err != nil {
		return nil, err // 缺少或获取 uid 失败
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return nil, err // uid 格式非法
	}

	sidStr, err := jwtlib.Get[string](token, "sid")
	if err != nil {
		return nil, err // 缺少或获取 sid 失败
	}
	sid, err := uuid.Parse(sidStr)
	if err != nil {
		return nil, err // sid 格式非法
	}

	name, _ := jwtlib.Get[string](token, "name")

	fp, _ := jwtlib.Get[string](token, "fp")

	return &Claims{
		Token:       token,
		UserUUID:    uid,
		SessionID:   sid,
		Name:        name,
		Fingerprint: fp,
	}, nil
}
