package conf

import "time"

type HTTP struct {
	Port       string        `yaml:"port"`
	ShutdownTO time.Duration `yaml:"shutdown_to"`
}

type Config struct {
	HTTP HTTP `yaml:"http"`
}
