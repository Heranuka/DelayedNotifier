// Package transport provides HTTP server lifecycle management.
package transport

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"delay/internal/config"

	"github.com/wb-go/wbf/logger"
)

// Server wraps http.Server and manages its lifecycle.
type Server struct {
	cfg    *config.Config
	log    *logger.ZerologAdapter
	router http.Handler
	server *http.Server
}

// New creates a new HTTP server instance.
func New(
	cfg *config.Config,
	log *logger.ZerologAdapter,
	router http.Handler,
) *Server {
	return &Server{
		cfg:    cfg,
		log:    log,
		router: router,
	}
}

// Run starts the HTTP server and blocks until the context is canceled or a fatal error occurs.
func (s *Server) Run(ctx context.Context) error {
	addr := normalizeAddr(s.cfg.HTTP.Port)

	s.server = &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadTimeout:       s.cfg.HTTP.ReadTimeout,
		WriteTimeout:      s.cfg.HTTP.WriteTimeout,
		IdleTimeout:       s.cfg.HTTP.IdleTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	s.log.Infow("http server starting",
		"addr", addr,
		"read_timeout", s.cfg.HTTP.ReadTimeout,
		"write_timeout", s.cfg.HTTP.WriteTimeout,
		"idle_timeout", s.cfg.HTTP.IdleTimeout,
	)

	errCh := make(chan error, 1)

	go func() {
		err := s.server.Serve(ln)
		if !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		s.log.Info("http server shutdown initiated")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.HTTP.ShutdownTimeout)
		defer cancel()

		if err := s.server.Shutdown(shutdownCtx); err != nil {
			s.log.Errorw("http server graceful shutdown failed", "error", err)

			if closeErr := s.server.Close(); closeErr != nil {
				s.log.Errorw("http server forced close failed", "error", closeErr)
			}

			return err
		}

		s.log.Info("http server stopped gracefully")
		return nil

	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("http server crashed: %w", err)
		}
		return nil
	}
}

// Shutdown stops the server gracefully.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	s.log.Info("http server shutting down")
	return s.server.Shutdown(ctx)
}

func normalizeAddr(port string) string {
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}
