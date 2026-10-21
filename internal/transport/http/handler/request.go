package handler

import (
	"budeku/iam/pkg/util"
	"net/http"
)

type RequestMeta struct {
	IP        string `json:"-"`
	UserAgent string `json:"-"`
}

func GetRequestMeta(r *http.Request) RequestMeta {
	return RequestMeta{
		IP:        util.ClientIP(r),
		UserAgent: util.UserAgent(r),
	}
}
