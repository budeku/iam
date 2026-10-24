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
	"budeku/iam/pkg/util"

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

	if !util.IsValidUsername(body.Username) || !util.IsValidPassword(body.Password) {
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

type UserListReq struct {
	Keyword string `form:"keyword,omitempty" binding:"max=32"`
	Pagination
}

func (h *UserHandler) List(c *gin.Context) {
	var params UserListReq
	if err := c.ShouldBindQuery(&params); err != nil {
		response.InvalidParams(c)
		return
	}

	list, count, err := h.userBiz.List(c.Request.Context(), params.Keyword, params.Page, params.PerPage)
	if err != nil {
		response.BadRequest(c)
		return
	}

	response.Success(c, ListResp{
		Content: list,
		Count:   count,
	})
}

func (h *UserHandler) Info(c *gin.Context) {
	var params InfoPathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.InvalidParams(c)
		return
	}

	info, err := h.userBiz.Info(c.Request.Context(), params.ID)
	if err != nil {
		response.BadRequest(c)
		return
	}

	response.Success(c, info)
}

type UserCreateReq struct {
	Username string `json:"username" binding:"required,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=18"`
}

func (h *UserHandler) Create(c *gin.Context) {
	var body UserCreateReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.InvalidParams(c)
		return
	}

	if !util.IsValidUsername(body.Username) ||
		!util.IsValidEmail(body.Email) ||
		!util.IsValidPassword(body.Password) {
		response.InvalidParams(c)
		return
	}

	_, err := h.userBiz.Create(
		c.Request.Context(),
		body.Username,
		body.Email,
		body.Password,
	)
	if err != nil {
		response.BadRequest(c)
		return
	}

	response.OK(c)
}

type UserUpdatePathParams struct {
	UserID int `uri:"id" binding:"required"`
}

type UserUpdateReq struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,max=32"`
	Status   bool   `json:"status"`
	Password string `json:"password" binding:"omitempty,min=6,max=18"`
}

func (h *UserHandler) Update(c *gin.Context) {
	var params UserUpdatePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	var body UserUpdateReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Err(c, code.InvalidParams)
		return
	}

	if !util.IsValidUsername(body.Username) || !util.IsValidEmail(body.Email) {
		response.Err(c, code.InvalidParams)
		return
	}

	if body.Password != "" && !util.IsValidPassword(body.Password) {
		response.Err(c, code.InvalidParams)
		return
	}

	err := h.userBiz.Update(c.Request.Context(), params.UserID, body.Username, body.Email, body.Password, body.Status)
	if err != nil {
		response.ErrMessage(c, code.BadRequest, err.Error())
		return
	}

	response.OK(c)
}

type UserUpdateStatusReq struct {
	Status bool `json:"status"`
}

func (h *UserHandler) UpdateStatus(c *gin.Context) {
	var params UserUpdatePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.InvalidParams(c)
		return
	}

	var body AppUpdateStatusReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.InvalidParams(c)
		return
	}

	if err := h.userBiz.UpdateStatus(c.Request.Context(), params.UserID, body.Status); err != nil {
		response.BadRequest(c)
		return
	}

	response.OK(c)
}

type UserUpdatePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required,min=6,max=18"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=18"`
}

func (h *UserHandler) UpdatePassword(c *gin.Context) {
	userUUID, ok := middleware.UserUUIDFromContext(c)
	if !ok {
		response.Err(c, code.Unauthorized)
		return
	}

	ctx := c.Request.Context()

	info, err := h.userBiz.GetByUUID(ctx, userUUID)
	if err != nil {
		response.Err(c, code.Unauthorized)
		return
	}

	var body UserUpdatePasswordReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.InvalidParams(c)
		return
	}

	if !util.IsValidPassword(body.OldPassword) || !util.IsValidPassword(body.NewPassword) {
		response.Err(c, code.InvalidParams)
		return
	}

	err = h.userBiz.UpdatePassword(ctx, info.ID, body.OldPassword, body.NewPassword)
	if err != nil {
		response.BadRequest(c)
		return
	}

	response.OK(c)
}

func (h *UserHandler) Delete(c *gin.Context) {
	var params DeletePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.InvalidParams(c)
		return
	}

	if err := h.userBiz.Delete(c.Request.Context(), params.ID); err != nil {
		response.BadRequest(c)
		return
	}

	response.OK(c)
}

func (h *UserHandler) Session(c *gin.Context) {
	userUUID, ok := middleware.UserUUIDFromContext(c)
	if !ok {
		response.Err(c, code.Unauthorized)
		return
	}
	currentSessionID, _ := middleware.SessionIDFromContext(c)

	ctx := c.Request.Context()

	info, err := h.userBiz.GetByUUID(ctx, userUUID)
	if err != nil {
		response.Err(c, code.Unauthorized)
		return
	}

	res, err := h.userBiz.Session(ctx, info.ID, currentSessionID)
	if err != nil {
		response.BadRequest(c)
		return
	}

	response.Success(c, res)
}

type UserRevokedSessionPathParams struct {
	SessionID int `uri:"id" binding:"required"`
}

func (h *UserHandler) RevokeSession(c *gin.Context) {
	var params UserRevokedSessionPathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.InvalidParams(c)
		return
	}

	userUUID, ok := middleware.UserUUIDFromContext(c)
	if !ok {
		response.Err(c, code.Unauthorized)
		return
	}

	ctx := c.Request.Context()

	u, err := h.userBiz.GetByUUID(ctx, userUUID)
	if err != nil {
		response.Err(c, code.Unauthorized)
		return
	}

	err = h.userBiz.RevokeSession(ctx, u.ID, params.SessionID)
	if err != nil {
		response.BadRequest(c)
		return
	}

	response.OK(c)
}
