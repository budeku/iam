package handler

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/conf"
	"budeku/iam/internal/consts"
	"budeku/iam/internal/logger"
	"budeku/iam/internal/transport/http/cookieutil"
	"budeku/iam/internal/transport/http/middleware"
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
		response.BadRequest(c)
		return
	}

	jwtCfg := conf.Get().JWT

	cookie := cookieutil.Cookie()
	cookie.Set(c.Writer, consts.CookieAccessToken, result.AccessToken, int(jwtCfg.AccessTokenTTL.Seconds()))
	cookie.Set(c.Writer, consts.CookieRefreshToken, result.RefreshToken, int(jwtCfg.RefreshTokenTTL.Seconds()))
	response.OK(c)
}

func (h *UserHandler) Logout(c *gin.Context) {
	cookie := cookieutil.Cookie()
	ctx := c.Request.Context()

	if sessionID, ok := middleware.SessionIDFromContext(c); ok {
		if err := h.userBiz.Logout(ctx, sessionID); err != nil {
			logger.Error("会话 %s 退出失败：%v", sessionID, err)
		}
	}

	cookie.Delete(c.Writer, consts.CookieAccessToken)
	cookie.Delete(c.Writer, consts.CookieRefreshToken)
	response.OK(c)
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := cookieutil.Cookie().Get(c.Request, consts.CookieRefreshToken)
	if err != nil || refreshToken == "" {
		response.Err(c, code.RefreshTokenInvalid)
		return
	}

	ctx := c.Request.Context()

	result, err := h.userBiz.RefreshToken(ctx, refreshToken)
	if err != nil {
		cookie := cookieutil.Cookie()
		cookie.Delete(c.Writer, consts.CookieAccessToken)
		cookie.Delete(c.Writer, consts.CookieRefreshToken)

		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	jwtCfg := conf.Get().JWT
	cookie := cookieutil.Cookie()
	cookie.Set(c.Writer, consts.CookieAccessToken, result.AccessToken, int(jwtCfg.AccessTokenTTL.Seconds()))
	cookie.Set(c.Writer, consts.CookieRefreshToken, result.RefreshToken, int(jwtCfg.RefreshTokenTTL.Seconds()))

	response.OK(c)
}

func (h *UserHandler) Me(c *gin.Context) {
	userUUID, ok := middleware.UserUUIDFromContext(c)
	if !ok {
		response.Err(c, code.Unauthorized)
		return
	}

	res, err := h.userBiz.Me(c.Request.Context(), userUUID)
	if err != nil {
		response.BadRequest(c)
		return
	}

	response.Success(c, res)
}
