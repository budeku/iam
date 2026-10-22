package jwt

import (
	"budeku/iam/pkg/hashutil"
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v4/jwa"
	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/lestrrat-go/jwx/v4/jws"
	jwtlib "github.com/lestrrat-go/jwx/v4/jwt"
)

var (
	ErrInvalidToken        = errors.New("jwt 无效令牌")
	ErrFingerprintMismatch = errors.New("jwt 客户端环境指纹不匹配")
)

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

func (m *Manager) SignAccessToken(ctx context.Context, userUUID, sessionID uuid.UUID, name string, appID int, ttl time.Duration, ip, userAgent string) (string, error) {
	return m.sign(ctx, userUUID, sessionID, name, appID, ttl, ip, userAgent)
}

func (m *Manager) sign(ctx context.Context, userUUID, sessionID uuid.UUID, name string, appID int, ttl time.Duration, ip, userAgent string) (string, error) {
	kid, priv, err := m.keys.SigningKey(ctx, appID)
	if err != nil {
		return "", err // 获取签名私钥失败
	}

	// 使用 kid 作为盐的一部分，保证即使密钥轮换，指纹算法的 salt 也匹配
	secretBytes := []byte(kid + "_fp_salt")
	fp := generateFingerprint(ip, userAgent, secretBytes)

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
		Claim("sid", sessionID.String()).
		Claim("fp", fp)

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

func (m *Manager) Parse(ctx context.Context, tokenString string, currentIP, currentUserAgent string) (*Claims, error) {
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

	claims, err := ParseClaims(token)
	if err != nil {
		return nil, err
	}

	// 校验环境指纹 (IP + UA)
	// 还原当时签名时用的动态 secret（利用 kid）
	secretBytes := []byte(kid + "_fp_salt")
	expectedFP := generateFingerprint(currentIP, currentUserAgent, secretBytes)

	if claims.Fingerprint != "" && claims.Fingerprint != expectedFP {
		return nil, ErrFingerprintMismatch
	}

	return claims, nil
}

func generateFingerprint(ip, userAgent string, secret []byte) string {
	// 模糊化：例如 "192.168.1.100" -> "192.168.1.0"
	maskedIP := maskIPv4(ip)

	// 拼接
	raw := maskedIP + "|" + userAgent

	// 截取哈希前 16 位减小 Token 体积
	hexStr, err := hashutil.MAC256Hex([]byte(raw), secret)
	if err != nil || len(hexStr) < 16 {
		// 降级兜底方案
		return hashutil.Sum256Hex([]byte(raw))[:16]
	}
	return hexStr[:16]
}

func maskIPv4(ip string) string {
	// 去除端口号（如果带了端口）
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		// 需排除 IPv6 的冒号情况，简单判断
		if !strings.Contains(ip, "]") && strings.Count(ip, ":") == 1 {
			ip = ip[:idx]
		}
	}
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return fmt.Sprintf("%s.%s.%s.0", parts[0], parts[1], parts[2])
	}
	return ip // 如果是 IPv6 或不规范格式，暂时原样返回
}
