package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"zmeow/internal/app/config"
	"zmeow/pkg/logger"
)

// Server representa o servidor HTTP da aplicação
type Server struct {
	httpServer *http.Server
	logger     logger.Logger
	config     *config.Config
}

// New cria uma nova instância do servidor
func New(cfg *config.Config, handler http.Handler, log logger.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%s", cfg.App.Host, cfg.App.Port),
			Handler:      handler,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		logger: log,
		config: cfg,
	}
}

// Start inicia o servidor HTTP
func (s *Server) Start() error {
	s.logger.WithField("addr", s.httpServer.Addr).Info().Msg("Starting HTTP server")
	
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	
	return nil
}

// Stop para o servidor HTTP graciosamente
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info().Msg("Shutting down HTTP server...")
	
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error().Msg("Server forced to shutdown")
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	
	s.logger.Info().Msg("HTTP server stopped gracefully")
	return nil
}