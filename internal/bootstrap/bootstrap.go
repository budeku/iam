package bootstrap

import (
	"budeku/iam/internal/conf"
	"budeku/iam/internal/infra/database"
	"budeku/iam/internal/infra/redis"
	"budeku/iam/internal/logger"
	"budeku/iam/internal/store"
	httpserver "budeku/iam/internal/transport/http"
	"budeku/iam/internal/transport/http/cookieutil"
	"budeku/iam/internal/wire"
	"budeku/iam/pkg/password"
	"budeku/iam/pkg/util"
	"context"
	"log"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

func clean(ctx context.Context, interval time.Duration) {
	sessionStore := store.NewSessionStore(database.DB())

	go func() {
		// 用于防止上一次清理还没结束，下一次 ticker 又触发导致任务重叠
		var running atomic.Int32

		cleanup := func() {
			// 如果上一次清理还没跑完，直接跳过本次，避免并发冲突
			if !running.CompareAndSwap(0, 1) {
				logger.Warn("上一轮会话清理尚未完成，跳过本次触发")
				return
			}
			defer running.Store(0)

			runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()

			total, err := sessionStore.ClearExpireOrRevokedAll(runCtx)
			if err != nil {
				logger.Error("清理过期/撤销会话失败: %s", err)
				return
			}
			if total > 0 {
				logger.Info("会话清理完成，清理数量：%d", total)
			}
		}

		// 启动时立即执行一次
		cleanup()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Info("会话清理调度器已停止")
				return
			case <-ticker.C:
				cleanup()
			}
		}
	}()
}

func Setup() {
	if err := conf.Init(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	cfg := conf.Get()

	if len(cfg.HTTP.TrustedProxies) > 0 {
		util.SetTrustedProxies(cfg.HTTP.TrustedProxies)
	}

	cookieutil.InitCookieUtil(cfg.Cookie)

	if err := logger.Init(cfg.Logger); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	password.Init(&password.Config{
		Time:           cfg.Password.Time,
		Memory:         cfg.Password.Memory,
		Threads:        cfg.Password.Threads,
		KeyLength:      cfg.Password.KeyLength,
		SaltLen:        cfg.Password.SaltLen,
		MaxConcurrency: cfg.Password.MaxConcurrency,
		WaitTimeout:    cfg.Password.WaitTimeout,
	})

	if err := database.Init(cfg.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := database.Migrate(ctx); err != nil {
		log.Fatalf("执行数据库迁移失败: %v", err)
	}

	if err := redis.Init(cfg.Redis); err != nil {
		log.Fatalf("初始化 Redis 失败: %v", err)
	}

	logger.Info("debug=%t", conf.Debug)
}

func Release() {
	if err := database.Close(); err != nil {
		logger.Error("关闭数据库失败: %s", err)
	}

	if err := redis.Close(); err != nil {
		logger.Error("关闭 Redis 失败: %s", err)
	}

	logger.Info("正在释放资源")
	_ = logger.Sync()
}

func Run() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	clean(ctx, 1*time.Hour)

	app := wire.InitApp(database.DB())
	srv, err := httpserver.New(app)
	if err != nil {
		logger.Fatal("创建 HTTP 服务失败: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- err
		}
	}()

	logger.Info("启动成功，Ctrl + C 停止")

	select {
	case <-ctx.Done():
	case err := <-errCh:
		stop()
		logger.Fatal("运行异常: %v", err)
	}

	logger.Info("正在关闭...")

	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Error("关闭出错: %v", err)
	}

	logger.Info("已关闭！")
}
