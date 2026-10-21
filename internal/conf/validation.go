package conf

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func validate(cfg *Config) (*Config, error) {
	var errs []error

	// HTTP
	if err := validateHTTP(cfg.HTTP); err != nil {
		errs = append(errs, err)
	}

	// Logger
	if err := validateLogger(cfg.Logger); err != nil {
		errs = append(errs, err)
	}

	// Database
	if err := validateDatabase(cfg.Database); err != nil {
		errs = append(errs, err)
	}

	// Password
	if err := validatePassword(cfg.Password); err != nil {
		errs = append(errs, err)
	}

	// Redis
	if err := validateRedis(cfg.Redis); err != nil {
		errs = append(errs, err)
	}

	// Cookie
	if err := validateCookie(cfg.Cookie); err != nil {
		errs = append(errs, err)
	}

	// JWT
	if err := validateJWT(cfg.JWT); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return cfg, nil
}

func validateHTTP(h HTTP) error {
	var errs []error
	if h.Port == "" {
		errs = append(errs, errors.New("http.port 不能为空"))
	} else {
		if !strings.HasPrefix(h.Port, ":") && !strings.Contains(h.Port, ":") {
			errs = append(errs, fmt.Errorf("http.port 格式错误，必须以冒号开头或包含 IP 地址，当前值: %s", h.Port))
		} else {
			portStr := strings.TrimPrefix(h.Port, ":")
			// 包含 IP
			if strings.Contains(h.Port, ":") && strings.Count(h.Port, ":") == 1 && !strings.HasPrefix(h.Port, ":") {
				_, pStr, err := net.SplitHostPort(h.Port)
				if err == nil {
					portStr = pStr
				}
			}

			if p, err := strconv.Atoi(portStr); err != nil || p < 1 || p > 65535 {
				errs = append(errs, fmt.Errorf("http.port 端口范围不合法 (1-65535): %s", h.Port))
			}
		}
	}
	if h.MaxConnections < 0 {
		errs = append(errs, errors.New("http.max_connections 不能为负数"))
	}
	if h.ReadTimeout < 0 {
		errs = append(errs, errors.New("http.read_timeout 不能为负数"))
	}
	if h.ReadHeaderTimeout < 0 {
		errs = append(errs, errors.New("http.read_header_timeout 不能为负数"))
	}
	if h.WriteTimeout < 0 {
		errs = append(errs, errors.New("http.write_timeout 不能为负数"))
	}
	if h.IdleTimeout < 0 {
		errs = append(errs, errors.New("http.idle_timeout 不能为负数"))
	}
	if h.MaxHeaderBytes < 0 {
		errs = append(errs, errors.New("http.max_header_bytes 不能为负数"))
	}
	if h.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("http.shutdown_timeout 必须大于 0"))
	}
	for i, proxy := range h.TrustedProxies {
		if _, _, err := net.ParseCIDR(proxy); err != nil {
			if parsedIP := net.ParseIP(proxy); parsedIP == nil {
				errs = append(errs, fmt.Errorf("http.trusted_proxies[%d] 不是合法的 IP 或 CIDR: %s", i, proxy))
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func validateLogger(l Logger) error {
	var errs []error
	if l.Filename == "" {
		errs = append(errs, errors.New("logger.filename 日志文件路径不能为空"))
	}
	if l.MaxSizeMB <= 0 {
		errs = append(errs, errors.New("logger.max_size_mb 日志文件最大尺寸 (MB) 必须大于 0"))
	}
	if l.MaxBackups < 0 {
		errs = append(errs, errors.New("logger.max_backups 日志最大备份数不能为负数"))
	}
	if l.MaxAgeDays < 0 {
		errs = append(errs, errors.New("logger.max_age_days 日志最大保存天数不能为负数"))
	}
	switch l.Level {
	case "debug", "info", "warn", "error":
		// 有效的日志级别
	default:
		errs = append(errs, fmt.Errorf("无效的日志级别 %q (允许的值: debug, info, warn, error)", l.Level))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func validateDatabase(d Database) error {
	var errs []error
	if d.Host == "" {
		errs = append(errs, errors.New("database.host 数据库地址不能为空"))
	}
	if d.Port == "" {
		errs = append(errs, errors.New("database.port 数据库端口不能为空"))
	}
	if d.User == "" {
		errs = append(errs, errors.New("database.user 数据库用户名不能为空"))
	}
	if d.DBName == "" {
		errs = append(errs, errors.New("database.db_name 数据库名称不能为空"))
	}
	if d.MaxIdleConns < 0 {
		errs = append(errs, errors.New("database.max_idle_conns 数据库最大空闲连接数不能为负数"))
	}
	if d.MaxOpenConns <= 0 {
		errs = append(errs, errors.New("database.max_open_conns 数据库最大打开连接数必须大于 0"))
	}
	if d.MaxIdleConns > d.MaxOpenConns {
		errs = append(errs, errors.New("数据库最大空闲连接数不能超过最大打开连接数"))
	}
	if d.MaxLifetime < 0 {
		errs = append(errs, errors.New("database.max_lifetime 数据库连接最大生命周期不能为负数"))
	}
	if d.MaxIdleTime < 0 {
		errs = append(errs, errors.New("database.max_idle_time 数据库连接最大空闲时间不能为负数"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func validatePassword(p Password) error {
	var errs []error
	if p.Time == 0 {
		errs = append(errs, errors.New("password.time 必须大于 0"))
	}
	if p.Memory == 0 {
		errs = append(errs, errors.New("password.memory 必须大于 0"))
	}
	if p.Threads == 0 {
		errs = append(errs, errors.New("password.threads 必须大于 0"))
	}
	if p.KeyLength == 0 {
		errs = append(errs, errors.New("password.key_length 必须大于 0"))
	}
	if p.SaltLen == 0 {
		errs = append(errs, errors.New("password.salt_len 必须大于 0"))
	}
	if p.MaxConcurrency <= 0 {
		errs = append(errs, errors.New("password.max_concurrency 必须大于 0"))
	}
	if p.WaitTimeout <= 0 {
		errs = append(errs, errors.New("password.wait_timeout 必须大于 0"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func validateRedis(r Redis) error {
	var errs []error

	if r.Addr == "" {
		errs = append(errs, errors.New("redis.addr 地址不能为空"))
	}
	if r.DB < 0 {
		errs = append(errs, errors.New("redis.db 不能为负数"))
	}
	if r.PoolSize <= 0 {
		errs = append(errs, errors.New("redis.pool_size 连接池大小必须大于 0"))
	}
	if r.MinIdleConns < 0 {
		errs = append(errs, errors.New("redis.min_idle_conns 最小空闲连接数不能为负数"))
	}
	if r.MinIdleConns > r.PoolSize {
		errs = append(errs, errors.New("最小空闲连接数不能超过连接池大小"))
	}
	if r.MaxIdleConns < 0 {
		errs = append(errs, errors.New("redis.max_idle_conns 最大空闲连接数不能为负数"))
	}
	if r.MaxIdleConns > 0 && r.MaxIdleConns > r.PoolSize {
		errs = append(errs, errors.New("最大空闲连接数不能超过连接池大小"))
	}
	if r.MaxIdleConns > 0 && r.MinIdleConns > r.MaxIdleConns {
		errs = append(errs, errors.New("最小空闲连接数不能超过最大空闲连接数"))
	}
	if r.MaxRetries < 0 {
		errs = append(errs, errors.New("redis.max_retries 最大重试次数不能为负数"))
	}
	if r.DialTimeout <= 0 {
		errs = append(errs, errors.New("redis.dial_timeout 连接超时时间必须大于 0"))
	}
	if r.ReadTimeout <= 0 {
		errs = append(errs, errors.New("redis.read_timeout 读取超时时间必须大于 0"))
	}
	if r.WriteTimeout <= 0 {
		errs = append(errs, errors.New("redis.write_timeout 写入超时时间必须大于 0"))
	}
	if r.PoolTimeout <= 0 {
		errs = append(errs, errors.New("redis.pool_timeout 池获取超时时间必须大于 0"))
	}
	if r.ConnMaxIdleTime < 0 {
		errs = append(errs, errors.New("redis.conn_max_idle_time 连接最大空闲时间不能为负数"))
	}
	if r.ConnMaxLifetime < 0 {
		errs = append(errs, errors.New("redis.conn_max_lifetime 连接最大生命周期不能为负数"))
	}
	if r.ConnMaxLifetime > 0 && r.ConnMaxIdleTime > r.ConnMaxLifetime {
		errs = append(errs, errors.New("redis 连接最大空闲时间不应超过最大生命周期"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func validateCookie(c Cookie) error {
	var errs []error

	if len(c.Secret) == 0 {
		errs = append(errs, errors.New("cookie.secret 不能为空"))
	}

	switch strings.ToLower(c.SameSite) {
	case "", "lax", "strict", "none":
	default:
		errs = append(errs, errors.New("cookie.samesite 必须是 'lax'、'strict' 或 'none' 之一"))
	}

	if strings.EqualFold(c.SameSite, "none") && !c.Secure {
		errs = append(errs, errors.New("当 cookie.samesite 为 'none' 时，cookie.secure 必须为 true"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func validateJWT(j JWT) error {
	var errs []error
	if j.Issuer == "" {
		errs = append(errs, errors.New("jwt.issuer 不能为空"))
	}
	if j.AccessTokenTTL <= 0 {
		errs = append(errs, errors.New("jwt.access_token_ttl 必须大于 0"))
	}
	if j.RefreshTokenTTL <= 0 {
		errs = append(errs, errors.New("jwt.refresh_token_ttl 必须大于 0"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
