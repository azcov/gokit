package server

import (
	"context"
	"errors"
	"net/http"
	"time"
)

type Config struct {
	Addr            string        `config:"addr"`
	ReadTimeout     time.Duration `config:"read_timeout"`
	WriteTimeout    time.Duration `config:"write_timeout"`
	IdleTimeout     time.Duration `config:"idle_timeout"`
	ShutdownTimeout time.Duration `config:"shutdown_timeout"`
}

type Server struct {
	srv *http.Server
	cfg Config
}

func New(cfg Config, handler http.Handler) *Server {
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 5 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 10 * time.Second
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 120 * time.Second
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = 30 * time.Second
	}
	return &Server{
		cfg: cfg,
		srv: &http.Server{
			Addr:         cfg.Addr,
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

// Start begins listening and serving. It blocks until the server stops.
func (s *Server) Start() error {
	if err := s.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully stops the server with a timeout.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

// Addr returns the configured listen address.
func (s *Server) Addr() string { return s.cfg.Addr }
