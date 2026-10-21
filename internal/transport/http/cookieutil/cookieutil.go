package cookieutil

import (
	"budeku/iam/internal/conf"
	"budeku/iam/pkg/hashutil"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidSignature = errors.New("cookie: 签名无效")
	ErrMalformedValue   = errors.New("cookie: 签名值格式错误")
)

var (
	defaultCookieUtil *CookieUtil
	cookieOnce        sync.Once
)

type CookieUtil struct {
	config conf.Cookie
}

type Option func(*http.Cookie)

func InitCookieUtil(cfg conf.Cookie) {
	cookieOnce.Do(func() {
		defaultCookieUtil = newCookieUtil(cfg)
	})
}

func Cookie() *CookieUtil {
	if defaultCookieUtil == nil {
		panic("cookieutil: cookie 工具尚未初始化，请先调用 InitCookieUtil")
	}
	return defaultCookieUtil
}

func newCookieUtil(cfg conf.Cookie) *CookieUtil {
	if cfg.Path == "" {
		cfg.Path = "/"
	}

	sameSite := parseSameSite(cfg.SameSite)
	if sameSite == http.SameSiteNoneMode {
		cfg.Secure = true
	}

	return &CookieUtil{config: cfg}
}

func parseSameSite(sameSite string) http.SameSite {
	switch strings.ToLower(sameSite) {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}

func WithDomain(domain string) Option {
	return func(c *http.Cookie) {
		c.Domain = domain
	}
}

func WithPath(path string) Option {
	return func(c *http.Cookie) {
		c.Path = path
	}
}

func WithSameSite(sameSite http.SameSite) Option {
	return func(c *http.Cookie) {
		c.SameSite = sameSite
	}
}

func WithHttpOnly(httpOnly bool) Option {
	return func(c *http.Cookie) {
		c.HttpOnly = httpOnly
	}
}

func (c *CookieUtil) Set(w http.ResponseWriter, name, value string, maxAge int, opts ...Option) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     c.config.Path,
		Domain:   c.config.Domain,
		MaxAge:   maxAge,
		HttpOnly: c.config.HttpOnly,
		Secure:   c.config.Secure,
		SameSite: parseSameSite(c.config.SameSite),
	}

	for _, opt := range opts {
		opt(cookie)
	}

	http.SetCookie(w, cookie)
}

func (c *CookieUtil) Get(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (c *CookieUtil) SetSigned(w http.ResponseWriter, name, value string, maxAge int, opts ...Option) error {
	if len(c.config.Secret) > 0 {
		sig, err := hashutil.MAC256Hex([]byte(value), []byte(c.config.Secret))
		if err != nil {
			return err
		}
		value = value + "." + sig
	}
	c.Set(w, name, value, maxAge, opts...)
	return nil
}

func (c *CookieUtil) GetSigned(r *http.Request, name string) (string, error) {
	val, err := c.Get(r, name)
	if err != nil {
		return "", err
	}

	parts := strings.Split(val, ".")
	if len(parts) != 2 {
		return "", ErrMalformedValue
	}

	value, sig := parts[0], parts[1]

	// 验证签名
	if !hashutil.VerifyMAC256Hex([]byte(value), []byte(c.config.Secret), sig) {
		return "", ErrInvalidSignature
	}

	return value, nil
}

func (c *CookieUtil) Delete(w http.ResponseWriter, name string, opts ...Option) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     c.config.Path,
		Domain:   c.config.Domain,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: c.config.HttpOnly,
		Secure:   c.config.Secure,
		SameSite: parseSameSite(c.config.SameSite),
	}

	for _, opt := range opts {
		opt(cookie)
	}

	http.SetCookie(w, cookie)
}
