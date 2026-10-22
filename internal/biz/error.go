package biz

import "errors"

var (
	ErrNotFound      = errors.New("信息不存在")
	ErrAlreadyExists = errors.New("信息已存在")
)
