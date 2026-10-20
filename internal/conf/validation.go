package conf

import "errors"

func validate(cfg *Config) (*Config, error) {
	var errs []error

	// HTTP
	if err := validateHTTP(cfg.HTTP); err != nil {
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
