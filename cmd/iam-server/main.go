package main

import (
	httpserver "budeku/iam/internal/transport/http"
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	srv, err := httpserver.New()
	if err != nil {
		log.Fatal(err)
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- err
		}
	}()

	fmt.Println("启动成功，Ctrl + C 停止")

	select {
	case <-ctx.Done():
	case err := <-errCh:
		stop()
		log.Fatal(err)
	}

	fmt.Println("正在关闭...")

	if err := srv.Shutdown(context.Background()); err != nil {
		fmt.Println(err)
	}

	fmt.Println("已关闭！")
}
