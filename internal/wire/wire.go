//go:build wireinject
// +build wireinject

package wire

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/conf"
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/jwt"
	"budeku/iam/internal/store"
	"budeku/iam/internal/transport/http/handler"

	"github.com/google/wire"
)

type App struct {
	User    *handler.UserHandler
	App     *handler.AppHandler
	Keypair *handler.KeypairHandler

	UserBiz    *biz.UserBiz
	SessionBiz *biz.SessionBiz

	JWTManager *jwt.Manager
	AuthToken  biz.AuthToken
}

var providerSet = wire.NewSet(
	store.NewAuthToken,
	wire.Bind(new(biz.AuthToken), new(*store.AuthToken)),

	store.NewTransactor,
	wire.Bind(new(biz.Transactor), new(*store.Transactor)),

	handler.NewKeyPairHandler,
	biz.NewKeypairBiz,
	store.NewKeypairStore,
	wire.Bind(new(biz.KeypairStore), new(*store.KeypairStore)),

	biz.NewSessionBiz,
	store.NewSessionStore,
	wire.Bind(new(biz.SessionStore), new(*store.SessionStore)),

	handler.NewUserHandler,
	biz.NewUserBiz,
	store.NewUserStore,
	wire.Bind(new(biz.UserStore), new(*store.UserStore)),

	biz.NewRefreshTokenBiz,
	store.NewRefreshTokenStore,
	wire.Bind(new(biz.RefreshTokenStore), new(*store.RefreshTokenStore)),

	handler.NewAppHandler,
	biz.NewAppBiz,
	store.NewAppStore,
	wire.Bind(new(biz.AppStore), new(*store.AppStore)),

	biz.NewBizKeyProvider,
	wire.Bind(new(jwt.KeyProvider), new(*biz.BizKeyProvider)),

	providerJWTManager,
	wire.Bind(new(biz.JwtSinger), new(*jwt.Manager)),

	wire.Struct(new(App), "*"),
)

func providerJWTManager(keyProvider *biz.BizKeyProvider) *jwt.Manager {
	cfg := conf.Get().JWT
	return jwt.New(keyProvider, jwt.WithIssuer(cfg.Issuer), jwt.WithAudience(cfg.Audience))
}

func InitApp(client *db.Client) *App {
	wire.Build(providerSet)
	return &App{}
}
