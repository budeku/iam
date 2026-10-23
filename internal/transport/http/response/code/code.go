package code

import "net/http"

type Code int

const (
	Success Code = 0 + iota
	Fail
)

const (
	BadRequest Code = 1000 + iota
	Unauthorized
	Forbidden
	NotFound
	Conflict
	TooManyRequests
	InternalError
)

const (
	InvalidParams Code = 2000 + iota
)

const (
	RefreshTokenInvalid Code = 3000 + iota
)

type meta struct {
	message    string
	httpStatus int
}

var registry = map[Code]meta{
	Success: {"success", http.StatusOK},
	Fail:    {"fail", http.StatusBadRequest},

	BadRequest:      {"错误请求", http.StatusBadRequest},
	Unauthorized:    {"用户未认证", http.StatusUnauthorized},
	Forbidden:       {"无权访问", http.StatusForbidden},
	NotFound:        {"资源不存在", http.StatusNotFound},
	Conflict:        {"资源冲突", http.StatusConflict},
	TooManyRequests: {"请求过于频繁", http.StatusTooManyRequests},
	InternalError:   {"服务器内部错误", http.StatusInternalServerError},

	InvalidParams: {"无效参数", http.StatusBadRequest},

	RefreshTokenInvalid: {"刷新令牌无效或过期", http.StatusUnauthorized},
}

func (c Code) Message() string {
	if m, ok := registry[c]; ok {
		return m.message
	}
	return "code 不存在"
}

func (c Code) HTTPStatus() int {
	if m, ok := registry[c]; ok {
		return m.httpStatus
	}
	return http.StatusInternalServerError
}

func (c Code) Int() int { return int(c) }
