package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/config"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/observability"
	"github.com/rs/zerolog"
)

// Service represents a service with lifecycle management
type Service struct {
	Name       string
	Version    string
	Config     *config.Config
	HTTPServer *http.Server
	logger     zerolog.Logger
}

// New creates a new service instance
func New(name, version string, cfg *config.Config) *Service {
	return &Service{
		Name:    name,
		Version: version,
		Config:  cfg,
		logger:  observability.GetLogger("service"),
	}
}

// Initialize initializes the service (logging, metrics, tracing)
func (s *Service) Initialize() error {
	// Initialize logging
	observability.InitLogging(
		s.Config.Logging.Level,
		s.Config.Logging.Format,
		s.Name,
	)

	s.logger = observability.GetLogger("service")
	observability.LogStartup(s.Name, s.Version, s.Config.Server.Port)

	// Initialize metrics if enabled
	if s.Config.Observability.MetricsEnabled {
		if err := observability.InitMetrics(s.Name); err != nil {
			return fmt.Errorf("failed to initialize metrics: %w", err)
		}
		s.logger.Info().Msg("Metrics initialized")
	}

	// Initialize tracing if enabled
	if s.Config.Observability.TracingEnabled {
		if err := observability.InitTracing(s.Name, s.Config.Observability.JaegerEndpoint); err != nil {
			s.logger.Warn().Err(err).Msg("Failed to initialize tracing, continuing without it")
		} else {
			s.logger.Info().Msg("Tracing initialized")
		}
	}

	return nil
}

// SetupHTTPServer sets up the HTTP server with the provided handler
func (s *Service) SetupHTTPServer(handler http.Handler) {
	s.HTTPServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.Config.Server.Host, s.Config.Server.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// Start starts the service and waits for shutdown signal
func (s *Service) Start() error {
	if s.HTTPServer == nil {
		return fmt.Errorf("HTTP server not configured")
	}

	// Start HTTP server in a goroutine
	go func() {
		s.logger.Info().
			Str("address", s.HTTPServer.Addr).
			Msg("Starting HTTP server")

		if err := s.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal().Err(err).Msg("HTTP server failed to start")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	s.waitForShutdown()

	return nil
}

// waitForShutdown waits for interrupt signal and performs graceful shutdown
func (s *Service) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal received
	sig := <-quit
	observability.LogShutdown(s.Name, fmt.Sprintf("received signal: %v", sig))

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if s.HTTPServer != nil {
		s.logger.Info().Msg("Shutting down HTTP server...")
		if err := s.HTTPServer.Shutdown(ctx); err != nil {
			s.logger.Error().Err(err).Msg("Failed to gracefully shutdown HTTP server")
		} else {
			s.logger.Info().Msg("HTTP server shutdown complete")
		}
	}

	s.logger.Info().Msg("Service shutdown complete")
}

// AddHealthCheck adds a health check endpoint to the provided router
func (s *Service) AddHealthCheck(mux *http.ServeMux) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"%s","version":"%s"}`, s.Name, s.Version)
	})
}

// AddMetricsEndpoint adds a metrics endpoint to the provided router
func (s *Service) AddMetricsEndpoint(mux *http.ServeMux) {
	if s.Config.Observability.MetricsEnabled {
		mux.Handle("/metrics", observability.GetMetricsHandler())
	}
}
