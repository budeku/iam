package conf

import (
	"fmt"
	"net/url"
	"time"
)

type HTTP struct {
	Port              string        `yaml:"port" mapstructure:"port"`
	MaxConnections    int           `yaml:"max_connections" mapstructure:"max_connections"`
	ReadTimeout       time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" mapstructure:"read_header_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout" mapstructure:"idle_timeout"`
	MaxHeaderBytes    int           `yaml:"max_header_bytes" mapstructure:"max_header_bytes"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout" mapstructure:"shutdown_timeout"`
	TrustedProxies    []string      `yaml:"trusted_proxies" mapstructure:"trusted_proxies"`
	CORSAllowOrigins  []string      `yaml:"cors_allow_origins" mapstructure:"cors_allow_origins"`
}

type Logger struct {
	Filename   string `yaml:"filename" mapstructure:"filename"`
	MaxSizeMB  int    `yaml:"max_size_mb" mapstructure:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups"`
	MaxAgeDays int    `yaml:"max_age_days" mapstructure:"max_age_days"`
	Compress   bool   `yaml:"compress" mapstructure:"compress"`
	Level      string `yaml:"level" mapstructure:"level"`
}

type Database struct {
	Host         string        `yaml:"host" mapstructure:"host"`
	Port         string        `yaml:"port" mapstructure:"port"`
	User         string        `yaml:"user" mapstructure:"user"`
	Password     string        `yaml:"password" mapstructure:"password"`
	DBName       string        `yaml:"db_name" mapstructure:"db_name"`
	Timezone     string        `yaml:"timezone" mapstructure:"timezone"`
	MaxIdleConns int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	MaxOpenConns int           `yaml:"max_open_conns" mapstructure:"max_open_conns"`
	MaxLifetime  time.Duration `yaml:"max_lifetime" mapstructure:"max_lifetime"`
	MaxIdleTime  time.Duration `yaml:"max_idle_time" mapstructure:"max_idle_time"`
}

func (d *Database) DSN() string {
	tz := d.Timezone
	if tz == "" {
		tz = "Local"
	}

	params := url.Values{}
	params.Set("charset", "utf8mb4")
	params.Set("parseTime", "True")
	params.Set("loc", tz)

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?%s",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.DBName,
		params.Encode(),
	)
}

type Password struct {
	Time           uint32        `yaml:"time" mapstructure:"time"`
	Memory         uint32        `yaml:"memory" mapstructure:"memory"`
	Threads        uint8         `yaml:"threads" mapstructure:"threads"`
	KeyLength      uint32        `yaml:"key_length" mapstructure:"key_length"`
	SaltLen        uint32        `yaml:"salt_len" mapstructure:"salt_len"`
	MaxConcurrency int           `yaml:"max_concurrency" mapstructure:"max_concurrency"`
	WaitTimeout    time.Duration `yaml:"wait_timeout" mapstructure:"wait_timeout"`
}

type Redis struct {
	Addr            string        `yaml:"addr" mapstructure:"addr"`
	Password        string        `yaml:"password" mapstructure:"password"`
	DB              int           `yaml:"db" mapstructure:"db"`
	Prefix          string        `yaml:"prefix" mapstructure:"prefix"`
	PoolSize        int           `yaml:"pool_size" mapstructure:"pool_size"`
	MinIdleConns    int           `yaml:"min_idle_conns" mapstructure:"min_idle_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	MaxRetries      int           `yaml:"max_retries" mapstructure:"max_retries"`
	DialTimeout     time.Duration `yaml:"dial_timeout" mapstructure:"dial_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	PoolTimeout     time.Duration `yaml:"pool_timeout" mapstructure:"pool_timeout"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" mapstructure:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
}

type Cookie struct {
	Domain   string `yaml:"domain" mapstructure:"domain"`
	Path     string `yaml:"path" mapstructure:"path"`
	Secure   bool   `yaml:"secure" mapstructure:"secure"`
	HttpOnly bool   `yaml:"http_only" mapstructure:"http_only"`
	SameSite string `yaml:"same_site" mapstructure:"same_site"` // lax, strict, none
	Secret   string `yaml:"secret" mapstructure:"secret"`
}

type JWT struct {
	Issuer          string        `yaml:"issuer" mapstructure:"issuer"`
	Audience        string        `yaml:"audience" mapstructure:"audience"`
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" mapstructure:"refresh_token_ttl"`
}

type Config struct {
	HTTP     HTTP     `yaml:"http" mapstructure:"http"`
	Logger   Logger   `yaml:"logger" mapstructure:"logger"`
	Database Database `yaml:"database" mapstructure:"database"`
	Password Password `yaml:"password" mapstructure:"password"`
	Redis    Redis    `yaml:"redis" mapstructure:"redis"`
	Cookie   Cookie   `yaml:"cookie" mapstructure:"cookie"`
	JWT      JWT      `yaml:"jwt" mapstructure:"jwt"`
}
