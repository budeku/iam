package handler

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/transport/http/response"

	"github.com/gin-gonic/gin"
)

type AppHandler struct {
	appBiz *biz.AppBiz
}

func NewAppHandler(appBiz *biz.AppBiz) *AppHandler {
	return &AppHandler{appBiz: appBiz}
}

type AppListReq struct {
	Keyword string `form:"keyword,omitempty" binding:"max=32"`
	Pagination
}

func (h *AppHandler) List(c *gin.Context) {
	var params AppListReq
	if err := c.ShouldBindQuery(&params); err != nil {
		response.InvalidParams(c)
		return
	}

	list, count, err := h.appBiz.List(c.Request.Context(), params.Keyword, params.Page, params.PerPage)
	if err != nil {
		response.BadRequest(c)
		return
	}

	response.Success(c, ListResp{
		Content: list,
		Count:   count,
	})
}

func (h *AppHandler) Info(c *gin.Context) {
	var params InfoPathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.InvalidParams(c)
		return
	}

	info, err := h.appBiz.Info(c.Request.Context(), params.ID)
	if err != nil {
		response.BadRequest(c)
		return
	}

	response.Success(c, info)
}

type AppCreateReq struct {
	Name         string   `json:"name" binding:"required,max=64"`
	ClientID     string   `json:"client_id" binding:"required,max=36"`
	RedirectUris []string `json:"redirect_uris" binding:"required,min=1,dive,required,url"`
	Status       bool     `json:"status"`
}

func (h *AppHandler) Create(c *gin.Context) {
	var body AppCreateReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.InvalidParams(c)
		return
	}

	_, err := h.appBiz.Create(
		c.Request.Context(),
		body.Name, body.ClientID,
		body.RedirectUris,
		body.Status,
	)
	if err != nil {
		response.BadRequest(c)
		return
	}

	response.OK(c)
}

type AppUpdatePathParams struct {
	AppID int `uri:"id" binding:"required"`
}

type AppUpdateReq struct {
	Name         string   `json:"name" binding:"required,max=64"`
	ClientID     string   `json:"client_id" binding:"required,max=36"`
	RedirectUris []string `json:"redirect_uris" binding:"required,min=1,dive,required,url"`
	Status       bool     `json:"status"`
}

func (h *AppHandler) Update(c *gin.Context) {
	var params AppUpdatePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.InvalidParams(c)
		return
	}

	var body AppUpdateReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.InvalidParams(c)
		return
	}

	if _, err := h.appBiz.Update(c.Request.Context(), params.AppID, body.Name, body.ClientID, body.RedirectUris, body.Status); err != nil {
		response.BadRequest(c)
		return
	}

	response.OK(c)
}

type AppUpdateStatusReq struct {
	Status bool `json:"status"`
}

func (h *AppHandler) UpdateStatus(c *gin.Context) {
	var params AppUpdatePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.InvalidParams(c)
		return
	}

	var body AppUpdateStatusReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.InvalidParams(c)
		return
	}

	if err := h.appBiz.UpdateStatus(c.Request.Context(), params.AppID, body.Status); err != nil {
		response.BadRequest(c)
		return
	}

	response.OK(c)
}

func (h *AppHandler) Delete(c *gin.Context) {
	var params DeletePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.InvalidParams(c)
		return
	}

	if err := h.appBiz.Delete(c.Request.Context(), params.ID); err != nil {
		response.BadRequest(c)
		return
	}

	response.OK(c)
}
