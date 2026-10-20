package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func New() *gin.Engine {
	r := gin.New()
	r.GET("/healthz", healthCheck)

	return r
}
