package util

import (
	"net"
	"net/url"
	"regexp"
	"sync"
)

var (
	usernameRegex *regexp.Regexp
	usernameOnce  sync.Once

	emailRegex *regexp.Regexp
	emailOnce  sync.Once

	passwordRegex *regexp.Regexp
	passwordOnce  sync.Once

	appNameRegex *regexp.Regexp
	appNameOnce  sync.Once
)

func getUsernameRegex() *regexp.Regexp {
	usernameOnce.Do(func() {
		usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{1,32}$`)
	})
	return usernameRegex
}

func getEmailRegex() *regexp.Regexp {
	emailOnce.Do(func() {
		emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	})
	return emailRegex
}

func getPasswordRegex() *regexp.Regexp {
	passwordOnce.Do(func() {
		passwordRegex = regexp.MustCompile(`^[^\s]{6,18}$`)
	})
	return passwordRegex
}

func getAppNameRegex() *regexp.Regexp {
	appNameOnce.Do(func() {
		appNameRegex = regexp.MustCompile(`^[\w\p{Han}]{1,64}$`)
	})
	return appNameRegex
}

func IsValidUsername(username string) bool {
	return getUsernameRegex().MatchString(username)
}

func IsValidEmail(email string) bool {
	if len(email) < 6 || len(email) > 128 {
		return false
	}
	return getEmailRegex().MatchString(email)
}

func IsValidPassword(password string) bool {
	return getPasswordRegex().MatchString(password)
}

func IsValidAppName(name string) bool {
	return getAppNameRegex().MatchString(name)
}

func IsValidIPOrURLList(list []string) bool {
	if len(list) == 0 {
		return false
	}

	for _, s := range list {
		isIP := net.ParseIP(s) != nil

		u, err := url.ParseRequestURI(s)
		isURL := err == nil && u.Scheme != "" && u.Host != ""

		if !isIP && !isURL {
			return false
		}
	}

	return true
}
