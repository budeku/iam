package middleware

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/consts"
	"budeku/iam/internal/jwt"
	"budeku/iam/internal/transport/http/cookieutil"
	"budeku/iam/internal/transport/http/response"
	"budeku/iam/internal/transport/http/response/code"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Auth(jwtManager *jwt.Manager, authToken biz.AuthToken) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := cookieutil.Cookie().Get(c.Request, consts.CookieAccessToken)
		if err != nil || accessToken == "" {
			response.Err(c, code.Unauthorized)
			c.Abort()
			return
		}

		claims, err := jwtManager.Parse(c.Request.Context(), accessToken)
		if err != nil {
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

		c.Set(consts.CtxUserUUID, claims.UserUUID)
		c.Set(consts.CtxSessionID, claims.SessionID)
		c.Set(consts.CtxUsername, claims.Name)

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
