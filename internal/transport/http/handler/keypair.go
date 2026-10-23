package handler

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/enum"
	"budeku/iam/internal/transport/http/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type KeypairHandler struct {
	keypairBiz *biz.KeypairBiz
}

func NewKeyPairHandler(keypairBiz *biz.KeypairBiz) *KeypairHandler {
	return &KeypairHandler{keypairBiz: keypairBiz}
}

type KeypairListReq struct {
	Pagination
}

type KeypairJWKSReq struct {
	ClientID uuid.UUID `form:"client_id" binding:"required"`
}

func (h *KeypairHandler) JWKS(c *gin.Context) {
	var params KeypairJWKSReq
	if err := c.ShouldBindQuery(&params); err != nil {
		response.InvalidParams(c)
		return
	}

	jwks, err := h.keypairBiz.JWKS(c.Request.Context(), params.ClientID)
	if err != nil {
		response.BadRequest(c)
		return
	}

	response.Success(c, jwks)
}

func (h *KeypairHandler) List(c *gin.Context) {
	var params KeypairListReq
	if err := c.ShouldBindQuery(&params); err != nil {
		response.InvalidParams(c)
		return
	}

	list, count, err := h.keypairBiz.List(c.Request.Context(), params.Page, params.PerPage)
	if err != nil {
		response.BadRequest(c)
		return
	}

	response.Success(c, ListResp{
		Content: list,
		Count:   count,
	})
}

type KeypaireCreateReq struct {
	Algorithm enum.KeypairAlgorithm `json:"algorithm" binding:"required"`
	AppID     int                   `json:"app_id" binding:"required"`
	Status    enum.KeypairStatus    `json:"status"`
}

func (h *KeypairHandler) Create(c *gin.Context) {
	var body KeypaireCreateReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.InvalidParams(c)
		return
	}

	if !body.Algorithm.Valid() || !body.Status.Valid() {
		response.BadRequest(c)
		return
	}

	_, err := h.keypairBiz.Create(c.Request.Context(), body.AppID, body.Algorithm, body.Status)
	if err != nil {
		response.BadRequest(c)
		return
	}

	response.OK(c)
}

func (h *KeypairHandler) Info(c *gin.Context) {
	var params InfoPathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.InvalidParams(c)
		return
	}

	info, err := h.keypairBiz.Info(c.Request.Context(), params.ID)
	if err != nil {
		response.BadRequest(c)
		return
	}

	response.Success(c, info)
}

type KeypairUpdatePathParams struct {
	KeypairID int `uri:"id" binding:"required"`
}

func (h *KeypairHandler) Downgrade(c *gin.Context) {
	var params KeypairUpdatePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.InvalidParams(c)
		return
	}

	if err := h.keypairBiz.Downgrade(c.Request.Context(), params.KeypairID); err != nil {
		response.BadRequest(c)
		return
	}

	response.OK(c)
}

func (h *KeypairHandler) Retire(c *gin.Context) {
	var params KeypairUpdatePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.InvalidParams(c)
		return
	}

	err := h.keypairBiz.Retire(c.Request.Context(), params.KeypairID)
	if err != nil {
		response.BadRequest(c)
		return
	}

	response.OK(c)
}
