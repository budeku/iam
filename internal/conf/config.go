package conf

import (
	"budeku/iam/pkg/random"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
)

const DataDir = "data"

var (
	once    sync.Once
	initErr error
	mu      sync.RWMutex
	current *Config
)

func defaultConfig() *Config {
	mustGenKey := func() string {
		if s, err := random.Alphanumeric(32); err == nil {
			return s
		}
		// openssl rand -base64 48 | tr -dc 'a-zA-Z0-9' | head -c 32
		// head -c 32 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c 32
		return "9dMO2doXM1p8wLep8xDnJT0xI6RITizE"
	}

	return &Config{
		HTTP: HTTP{
			Port:            ":26001",
			ShutdownTimeout: 30 * time.Second,
		},
		Logger: Logger{
			Filename:   filepath.Join(DataDir, "logs/iam.log"),
			MaxSizeMB:  100,
			MaxBackups: 7,
			MaxAgeDays: 30,
			Compress:   true,
			// debug < info < warn < error < fatal
			// info 时会记录到 info & warn & error & fatal
			Level: "info",
		},
		Database: Database{
			Host:         "127.0.0.1",
			Port:         "3306",
			User:         "root",
			Password:     "",
			DBName:       "iam",
			Timezone:     "Local",
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			MaxLifetime:  1 * time.Hour,
			MaxIdleTime:  5 * time.Minute,
		},
		Password: Password{
			Time:           3,
			Memory:         32 * 1024, // 32MB
			Threads:        2,
			KeyLength:      32,
			SaltLen:        16,
			MaxConcurrency: 8,
			WaitTimeout:    5 * time.Second,
		},
		Redis: Redis{
			Addr:            "127.0.0.1:6379",
			Password:        "",
			DB:              0,
			Prefix:          "iam",
			PoolSize:        100,
			MinIdleConns:    10,
			MaxIdleConns:    50,
			MaxRetries:      3,
			DialTimeout:     5 * time.Second,
			ReadTimeout:     3 * time.Second,
			WriteTimeout:    3 * time.Second,
			PoolTimeout:     4 * time.Second,
			ConnMaxIdleTime: 30 * time.Minute,
			ConnMaxLifetime: time.Hour,
		},
		Cookie: Cookie{
			Domain:   "",
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
			SameSite: "lax",
			Secret:   mustGenKey(),
		},
	}
}

func load() (*Config, error) {
	cfgP := filepath.Join(DataDir, "config.yaml")

	_, err := os.Stat(cfgP)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		// 不存在,写入默认配置
		dir := filepath.Dir(cfgP)
		if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
			return nil, mkErr
		}

		defaultCfg := defaultConfig()
		data, mErr := yaml.Marshal(defaultCfg)
		if mErr != nil {
			return nil, mErr
		}

		if wErr := os.WriteFile(cfgP, data, 0644); wErr != nil {
			return nil, wErr
		}

		return validate(defaultCfg)
	}

	data, err := os.ReadFile(cfgP)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return validate(&cfg)
}

func Init() error {
	once.Do(func() {
		if err := os.MkdirAll(DataDir, 0755); err != nil {
			initErr = err
			return
		}

		cfg, err := load()
		if err != nil {
			initErr = err
			return
		}

		mu.Lock()
		current = cfg
		mu.Unlock()
	})
	return initErr
}

func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	if current == nil {
		panic("配置未初始化，请先调用 Init()")
	}
	return current
}
