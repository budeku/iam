package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

func CORS(allowOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" && slices.Contains(allowOrigins, origin) {
			// 严格匹配白名单中的 Origin
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin") // 避免多域名 CDN 缓存混乱
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")

			// 动态回显客户端请求的 Headers，避免写死自定义 Header
			if reqHeaders := c.Request.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
				c.Header("Access-Control-Allow-Headers", reqHeaders)
			} else {
				c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			}

			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400") // 缓存预检请求 24 小时
		}

		// 处理 OPTIONS 预检请求
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
