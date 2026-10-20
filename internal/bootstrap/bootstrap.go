package bootstrap

import (
	"budeku/iam/internal/conf"
	"budeku/iam/internal/infra/database"
	"budeku/iam/internal/logger"
	httpserver "budeku/iam/internal/transport/http"
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func setup() {
	if err := conf.Init(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	cfg := conf.Get()
	if err := logger.Init(cfg.Logger); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	if err := database.Init(cfg.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := database.Migrate(ctx); err != nil {
		log.Fatalf("执行数据库迁移失败: %v", err)
	}

	logger.Info("debug=%t", conf.Debug)
}

func release() {
	if err := database.Close(); err != nil {
		logger.Error("关闭数据库失败: %s", err)
	}

	logger.Info("正在释放资源")
	_ = logger.Sync()
}

func Run() {
	setup()
	defer release()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	srv, err := httpserver.New()
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
