package httpserver

import (
	"context"
	"fmt"
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

func New(address string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    address,
			Handler: handler,
		},
	}
}

func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("start http server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
