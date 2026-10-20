package conf

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
)

const dataDir = "data"

var (
	once    sync.Once
	initErr error
	mu      sync.RWMutex
	current *Config
)

func defaultConfig() *Config {
	return &Config{
		HTTP: HTTP{
			Port:       ":26001",
			ShutdownTO: 30 * time.Second,
		},
		Logger: Logger{
			Filename:   filepath.Join(dataDir, "logs/iam.log"),
			MaxSizeMB:  100,
			MaxBackups: 7,
			MaxAgeDays: 30,
			Compress:   true,
			// debug < info < warn < error < fatal
			// info 时会记录到 info & warn & error & fatal
			Level: "info",
		},
	}
}

func load() (*Config, error) {
	cfgP := filepath.Join(dataDir, "config.yaml")

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
		if err := os.MkdirAll(dataDir, 0755); err != nil {
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
