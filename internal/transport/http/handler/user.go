package handler

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/conf"
	"budeku/iam/internal/consts"
	"budeku/iam/internal/transport/http/cookieutil"
	"budeku/iam/internal/transport/http/response"
	"budeku/iam/internal/transport/http/response/code"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userBiz *biz.UserBiz
}

func NewUserHandler(userBiz *biz.UserBiz) *UserHandler {
	return &UserHandler{userBiz: userBiz}
}

type LoginReq struct {
	Username    string      `json:"username" binding:"required,max=32"`
	Password    string      `json:"password" binding:"required,min=6,max=18"`
	RequestMeta RequestMeta `json:"-"`
}

func (h *UserHandler) Login(c *gin.Context) {
	var body LoginReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.InvalidParams(c)
		return
	}

	body.RequestMeta = GetRequestMeta(c.Request)

	ctx := c.Request.Context()

	result, err := h.userBiz.Login(ctx, body.Username, body.Password, body.RequestMeta.IP, body.RequestMeta.UserAgent)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	jwtCfg := conf.Get().JWT

	cookie := cookieutil.Cookie()
	cookie.Set(c.Writer, consts.CookieAccessToken, result.AccessToken, int(jwtCfg.AccessTokenTTL.Seconds()))
	cookie.Set(c.Writer, consts.CookieRefreshToken, result.RefreshToken, int(jwtCfg.RefreshTokenTTL.Seconds()))
	response.OK(c)
}
