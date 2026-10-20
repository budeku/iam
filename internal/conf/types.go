package conf

import "time"

type HTTP struct {
	Port       string        `yaml:"port"`
	ShutdownTO time.Duration `yaml:"shutdown_to"`
}

type Logger struct {
	Filename   string `yaml:"filename"`
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAgeDays int    `yaml:"max_age_days"`
	Compress   bool   `yaml:"compress"`
	Level      string `yaml:"level"`
}

type Config struct {
	HTTP   HTTP   `yaml:"http"`
	Logger Logger `yaml:"logger"`
}
