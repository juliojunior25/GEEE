package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yourusername/geee/internal/core"
	"github.com/yourusername/geee/pkg/types"
)

// ServerConfig holds the server configuration
type ServerConfig struct {
	// Port is the port to listen on
	Port int

	// Host is the host to bind to
	Host string

	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout time.Duration

	// EnableCORS enables CORS middleware
	EnableCORS bool

	// AllowedOrigins is a list of allowed origins for CORS
	AllowedOrigins []string

	// ShutdownTimeout is the maximum duration to wait for graceful shutdown
	ShutdownTimeout time.Duration
}

// DefaultServerConfig returns a default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:            8080,
		Host:            "0.0.0.0",
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		EnableCORS:      true,
		AllowedOrigins:  []string{"*"},
		ShutdownTimeout: 15 * time.Second,
	}
}

// Server represents the HTTP server
type Server struct {
	config   *ServerConfig
	handlers *Handlers
	logger   types.Logger
	metrics  types.MetricsCollector
	server   *http.Server
}

// NewServer creates a new HTTP server
func NewServer(
	config *ServerConfig,
	registry core.PluginRegistry,
	executor core.Executor,
	logger types.Logger,
	metrics types.MetricsCollector,
) *Server {
	if config == nil {
		config = DefaultServerConfig()
	}

	handlers := NewHandlers(registry, executor, logger)

	return &Server{
		config:   config,
		handlers: handlers,
		logger:   logger,
		metrics:  metrics,
	}
}

// setupRouter sets up the Chi router with all routes and middlewares
func (s *Server) setupRouter() *chi.Mux {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(RecoveryMiddleware(s.logger))
	r.Use(LoggingMiddleware(s.logger))
	r.Use(MetricsMiddleware(s.metrics))

	// CORS middleware (if enabled)
	if s.config.EnableCORS {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   s.config.AllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}))
	}

	// Content-Type middleware for JSON responses
	r.Use(ContentTypeMiddleware)

	// Routes
	r.Get("/health", s.handlers.HandleHealth)
	r.Get("/ready", s.handlers.HandleReady)
	r.Get("/plugins", s.handlers.HandlePlugins)
	r.Post("/run", s.handlers.HandleRun)

	// Metrics endpoint (Prometheus)
	r.Handle("/metrics", promhttp.Handler())

	return r
}

// Start starts the HTTP server
func (s *Server) Start() error {
	router := s.setupRouter()

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	s.logger.Info("server_starting", map[string]interface{}{
		"host": s.config.Host,
		"port": s.config.Port,
		"addr": addr,
	})

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return fmt.Errorf("server failed to start: %w", err)
	case sig := <-sigChan:
		s.logger.Info("shutdown_signal_received", map[string]interface{}{
			"signal": sig.String(),
		})
		return s.Shutdown()
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	if s.server == nil {
		return nil
	}

	s.logger.Info("server_shutting_down", map[string]interface{}{
		"timeout": s.config.ShutdownTimeout.String(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("shutdown_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.logger.Info("server_stopped", nil)
	return nil
}
