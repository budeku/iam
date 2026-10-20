package router

import (
	"budeku/iam/internal/conf"
	"net/http"

	"github.com/gin-gonic/gin"
)

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func New() *gin.Engine {
	if conf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	if conf.Debug {
		r.Use(gin.Logger(), gin.Recovery())
	}

	r.GET("/healthz", healthCheck)

	return r
}
