package jwt

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v4/jwa"
	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/lestrrat-go/jwx/v4/jws"
	jwtlib "github.com/lestrrat-go/jwx/v4/jwt"
)

var ErrInvalidToken = errors.New("jwt 无效令牌")

type Manager struct {
	keys     KeyProvider
	issuer   string
	audience string
	leeway   time.Duration
}

type Option func(*Manager)

func WithIssuer(issuer string) Option   { return func(m *Manager) { m.issuer = issuer } }
func WithAudience(aud string) Option    { return func(m *Manager) { m.audience = aud } }
func WithLeeway(d time.Duration) Option { return func(m *Manager) { m.leeway = d } }

func New(keys KeyProvider, opts ...Option) *Manager {
	m := &Manager{keys: keys}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Manager) SignAccessToken(ctx context.Context, userUUID, sessionID uuid.UUID, name string, appID int, ttl time.Duration) (string, error) {
	return m.sign(ctx, userUUID, sessionID, name, appID, ttl)
}

func (m *Manager) sign(ctx context.Context, userUUID, sessionID uuid.UUID, name string, appID int, ttl time.Duration) (string, error) {
	kid, priv, err := m.keys.SigningKey(ctx, appID)
	if err != nil {
		return "", err // 获取签名私钥失败
	}

	var alg jwa.SignatureAlgorithm
	switch priv.(type) {
	case *rsa.PrivateKey:
		alg = jwa.RS256()
	case *ecdsa.PrivateKey:
		alg = jwa.ES256()
	default:
		return "", fmt.Errorf("不支持的私钥类型：%T", priv)
	}

	jwkKey, err := jwk.Import[jwk.Key](priv)
	if err != nil {
		return "", err // 私钥转换为 JWK 失败
	}
	_ = jwkKey.Set(jwk.KeyIDKey, kid)

	now := time.Now()
	builder := jwtlib.NewBuilder().
		Issuer(m.issuer).
		IssuedAt(now).
		Expiration(now.Add(ttl)).
		Claim("uid", userUUID.String()).
		Claim("sid", sessionID.String())

	if m.audience != "" {
		builder.Audience([]string{m.audience})
	}
	if name != "" {
		builder.Claim("name", name)
	}

	token, err := builder.Build()
	if err != nil {
		return "", err // 构建 Token 失败
	}

	signed, err := jwtlib.Sign(token, jwtlib.WithKey(alg, jwkKey))
	if err != nil {
		return "", err // 签名失败
	}

	return string(signed), nil
}

func (m *Manager) Parse(ctx context.Context, tokenString string) (*Claims, error) {
	msg, err := jws.Parse([]byte(tokenString))
	if err != nil {
		return nil, err // 解析 JWS 结构失败
	}

	if len(msg.Signatures()) == 0 {
		return nil, ErrInvalidToken // 缺少签名信息
	}

	headers := msg.Signatures()[0].ProtectedHeaders()
	kid, _ := headers.KeyID()
	alg, _ := headers.Algorithm()

	pub, err := m.keys.VerifyKey(ctx, kid)
	if err != nil {
		return nil, err // 获取验证公钥失败
	}

	parseOpts := []jwtlib.ParseOption{
		jwtlib.WithVerify(true),
		jwtlib.WithKey(alg, pub),
		jwtlib.WithValidate(true),
	}

	if m.leeway > 0 {
		parseOpts = append(parseOpts, jwtlib.WithAcceptableSkew(m.leeway))
	}
	if m.issuer != "" {
		parseOpts = append(parseOpts, jwtlib.WithIssuer(m.issuer))
	}
	if m.audience != "" {
		parseOpts = append(parseOpts, jwtlib.WithAudience(m.audience))
	}

	token, err := jwtlib.Parse([]byte(tokenString), parseOpts...)
	if err != nil {
		return nil, err // Token 验签或校验失败（过期等）
	}

	return ParseClaims(token)
}
