package conf

import (
	"fmt"
	"net/url"
	"time"
)

type HTTP struct {
	Port            string        `yaml:"port"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type Logger struct {
	Filename   string `yaml:"filename"`
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAgeDays int    `yaml:"max_age_days"`
	Compress   bool   `yaml:"compress"`
	Level      string `yaml:"level"`
}

type Database struct {
	Host         string        `yaml:"host"`
	Port         string        `yaml:"port"`
	User         string        `yaml:"user"`
	Password     string        `yaml:"password"`
	DBName       string        `yaml:"db_name"`
	Timezone     string        `yaml:"timezone"`
	MaxIdleConns int           `yaml:"max_idle_conns"`
	MaxOpenConns int           `yaml:"max_open_conns"`
	MaxLifetime  time.Duration `yaml:"max_lifetime"`
	MaxIdleTime  time.Duration `yaml:"max_idle_time"`
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
	Time           uint32        `yaml:"time"`
	Memory         uint32        `yaml:"memory"`
	Threads        uint8         `yaml:"threads"`
	KeyLength      uint32        `yaml:"key_length"`
	SaltLen        uint32        `yaml:"salt_len"`
	MaxConcurrency int           `yaml:"max_concurrency"`
	WaitTimeout    time.Duration `yaml:"wait_timeout"`
}

type Redis struct {
	Addr            string        `yaml:"addr"`
	Password        string        `yaml:"password"`
	DB              int           `yaml:"db"`
	Prefix          string        `yaml:"prefix"`
	PoolSize        int           `yaml:"pool_size"`
	MinIdleConns    int           `yaml:"min_idle_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	MaxRetries      int           `yaml:"max_retries"`
	DialTimeout     time.Duration `yaml:"dial_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	PoolTimeout     time.Duration `yaml:"pool_timeout"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type Cookie struct {
	Domain   string `yaml:"domain"`
	Path     string `yaml:"path"`
	Secure   bool   `yaml:"secure"`
	HttpOnly bool   `yaml:"http_only"`
	SameSite string `yaml:"same_site"` // lax, strict, none
	Secret   string `yaml:"secret"`
}

type Config struct {
	HTTP     HTTP     `yaml:"http"`
	Logger   Logger   `yaml:"logger"`
	Database Database `yaml:"database"`
	Password Password `yaml:"password"`
	Redis    Redis    `yaml:"redis"`
	Cookie   Cookie   `yaml:"cookie"`
}
