package middleware

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/consts"
	"budeku/iam/internal/jwt"
	"budeku/iam/internal/logger"
	"budeku/iam/internal/transport/http/cookieutil"
	"budeku/iam/internal/transport/http/response"
	"budeku/iam/internal/transport/http/response/code"
	"budeku/iam/pkg/util"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Auth(jwtManager *jwt.Manager, authToken biz.AuthToken, sessionBiz *biz.SessionBiz) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := cookieutil.Cookie().Get(c.Request, consts.CookieAccessToken)
		if err != nil || accessToken == "" {
			response.Err(c, code.Unauthorized)
			c.Abort()
			return
		}

		claims, err := jwtManager.Parse(c.Request.Context(), accessToken, util.ClientIP(c.Request), util.UserAgent(c.Request))
		if err != nil {
			logger.Warn("JWT 校验失败: %v | 客户端 IP: %s | User-Agent: %s", err, util.ClientIP(c.Request), util.UserAgent(c.Request))
			response.Err(c, code.Unauthorized)
			c.Abort()
			return
		}

		online, err := authToken.ExistsAccess(c.Request.Context(), claims.SessionID)
		if err != nil || !online {
			response.Err(c, code.Unauthorized)
			c.Abort()
			return
		}

		sessionBiz.Touch(c.Request.Context(), claims.SessionID)

		c.Set(consts.CtxUserUUID, claims.UserUUID)
		c.Set(consts.CtxSessionID, claims.SessionID)
		c.Set(consts.CtxUsername, claims.Name)

		c.Next()
	}
}

func RequireSuperAdmin(userBiz *biz.UserBiz) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUUID, ok := UserUUIDFromContext(c)
		if !ok {
			response.Err(c, code.Unauthorized)
			c.Abort()
			return
		}

		ctx := c.Request.Context()

		info, err := userBiz.GetByUUID(ctx, userUUID)
		if err != nil {
			if errors.Is(err, biz.ErrNotFound) {
				response.Unauthorized(c)
			} else {
				response.InternalError(c)
			}
			c.Abort()
			return
		}

		if !info.IsSuperAdmin {
			response.Err(c, code.Forbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

func UserUUIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(consts.CtxUserUUID)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func SessionIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(consts.CtxSessionID)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}
