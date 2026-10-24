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
	r.GET("/.well-known/jwks.json", app.Keypair.JWKS)

	api := r.Group("/api")
	auth := api.Group("/auth")
	{
		auth.POST("/login", app.User.Login)
		auth.POST("/refresh-token", app.User.RefreshToken)
	}

	protected := api.Group("")
	protected.Use(middleware.Auth(app.JWTManager, app.AuthToken, app.SessionBiz))

	protected.POST("/auth/logout", app.User.Logout)
	protected.GET("/auth/me", app.User.Me)
	protected.GET("/users/sessions", app.User.Session)
	protected.DELETE("/users/sessions/:id", app.User.RevokeSession)
	protected.PUT("/users/password", app.User.UpdatePassword)

	admin := protected.Group("")
	admin.Use(middleware.RequireSuperAdmin(app.UserBiz))
	{
		admin.GET("/apps", app.App.List)
		admin.GET("/apps/options", app.App.Options)
		admin.GET("/apps/:id", app.App.Info)
		admin.POST("/apps", app.App.Create)
		admin.PUT("/apps/:id", app.App.Update)
		admin.PUT("/apps/:id/status", app.App.UpdateStatus)
		admin.DELETE("/apps/:id", app.App.Delete)

		admin.GET("/keypairs", app.Keypair.List)
		admin.POST("/keypairs", app.Keypair.Create)
		admin.GET("/keypairs/:id", app.Keypair.Info)
		admin.PUT("/keypairs/:id/downgrade", app.Keypair.Downgrade)
		admin.PUT("/keypairs/:id/retire", app.Keypair.Retire)
		admin.GET("/keypairs/status-options", app.Keypair.StatusOptions)
		admin.GET("/keypairs/algorithm-options", app.Keypair.AlgorithmOptions)

		admin.GET("/users", app.User.List)
		admin.GET("/users/:id", app.User.Info)
		admin.POST("/users", app.User.Create)
		admin.PUT("/users/:id", app.User.Update)
		admin.PUT("/users/:id/status", app.User.UpdateStatus)
		admin.DELETE("/users/:id", app.User.Delete)
	}
}
