package http

import (
	"budeku/iam/internal/conf"
	"budeku/iam/internal/transport/http/router"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
	listener   net.Listener
	shutdownTO time.Duration
}

func New() (*Server, error) {
	cfg := conf.Get()
	addr := cfg.HTTP.Port
	shutdownTO := cfg.HTTP.ShutdownTO

	engine := router.New()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("监听 %s 失败：%w", addr, err)
	}

	srv := &http.Server{
		Handler: engine,
	}

	return &Server{
		httpServer: srv,
		listener:   ln,
		shutdownTO: shutdownTO,
	}, nil
}

func (s *Server) Start() error {
	if err := s.httpServer.Serve(s.listener); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("启动失败：%w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, s.shutdownTO)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("关机失败：%w", err)
	}
	return nil
}
