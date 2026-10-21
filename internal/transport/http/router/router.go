package router

import (
	"budeku/iam/internal/conf"
	"budeku/iam/internal/transport/http/middleware"
	"budeku/iam/internal/wire"
	"net/http"

	"github.com/gin-gonic/gin"
)

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func New(app *wire.App) *gin.Engine {
	if conf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Recovery())

	if conf.Debug {
		r.Use(gin.Logger())
	} else {
		r.Use(middleware.RequestLogger())
	}

	r.GET("/healthz", healthCheck)

	registerRoutes(r, app)

	return r
}

func registerRoutes(r *gin.Engine, app *wire.App) {
	api := r.Group("/api")
	auth := api.Group("/auth")
	{
		auth.POST("/login", app.User.Login)
	}

	protected := api.Group("")
	protected.Use(middleware.Auth(app.JWTManager, app.AuthToken))

	protected.POST("/auth/logout", app.User.Logout)
	protected.GET("/auth/me", app.User.Me)
}
