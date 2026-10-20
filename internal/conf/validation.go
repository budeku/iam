package conf

import (
	"errors"
	"fmt"
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

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return cfg, nil
}

func validateHTTP(h HTTP) error {
	var errs []error
	if h.Port == "" {
		errs = append(errs, errors.New("http.port 不能为空"))
	}
	if h.ShutdownTO <= 0 {
		errs = append(errs, errors.New("http.shutdown_to 必须大于 0"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func validateLogger(l Logger) error {
	var errs []error
	if l.Filename == "" {
		errs = append(errs, errors.New("日志文件路径不能为空"))
	}
	if l.MaxSizeMB <= 0 {
		errs = append(errs, errors.New("日志文件最大尺寸 (MB) 必须大于 0"))
	}
	if l.MaxBackups < 0 {
		errs = append(errs, errors.New("日志最大备份数不能为负数"))
	}
	if l.MaxAgeDays < 0 {
		errs = append(errs, errors.New("日志最大保存天数不能为负数"))
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
		errs = append(errs, errors.New("数据库地址不能为空"))
	}
	if d.Port == "" {
		errs = append(errs, errors.New("数据库端口不能为空"))
	}
	if d.User == "" {
		errs = append(errs, errors.New("数据库用户名不能为空"))
	}
	if d.DBName == "" {
		errs = append(errs, errors.New("数据库名称不能为空"))
	}
	if d.MaxIdleConns < 0 {
		errs = append(errs, errors.New("数据库最大空闲连接数不能为负数"))
	}
	if d.MaxOpenConns <= 0 {
		errs = append(errs, errors.New("数据库最大打开连接数必须大于 0"))
	}
	if d.MaxIdleConns > d.MaxOpenConns {
		errs = append(errs, errors.New("数据库最大空闲连接数不能超过最大打开连接数"))
	}
	if d.MaxLifetime < 0 {
		errs = append(errs, errors.New("数据库连接最大生命周期不能为负数"))
	}
	if d.MaxIdleTime < 0 {
		errs = append(errs, errors.New("数据库连接最大空闲时间不能为负数"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
