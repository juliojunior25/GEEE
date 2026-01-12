package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yourusername/geee/internal/executor"
	"github.com/yourusername/geee/internal/observability"
	"github.com/yourusername/geee/internal/registry"
	"github.com/yourusername/geee/internal/server"
	"github.com/yourusername/geee/plugins"
)

const (
	defaultPort            = 8080
	defaultHost            = "0.0.0.0"
	defaultReadTimeout     = 30
	defaultWriteTimeout    = 30
	defaultShutdownTimeout = 15
)

func main() {
	// Parse flags
	port := flag.Int("port", defaultPort, "Port to listen on")
	host := flag.String("host", defaultHost, "Host to bind to")
	readTimeout := flag.Int("read-timeout", defaultReadTimeout, "Read timeout in seconds")
	writeTimeout := flag.Int("write-timeout", defaultWriteTimeout, "Write timeout in seconds")
	shutdownTimeout := flag.Int("shutdown-timeout", defaultShutdownTimeout, "Shutdown timeout in seconds")
	enableCORS := flag.Bool("cors", true, "Enable CORS middleware")
	allowedOrigins := flag.String("allowed-origins", "*", "Comma-separated list of allowed origins for CORS")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")

	flag.Parse()

	// Check environment variables (they override flags)
	if envPort := os.Getenv("GEEE_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			*port = p
		}
	}
	if envHost := os.Getenv("GEEE_HOST"); envHost != "" {
		*host = envHost
	}
	if envCORS := os.Getenv("GEEE_ENABLE_CORS"); envCORS != "" {
		*enableCORS = envCORS == "true" || envCORS == "1"
	}
	if envOrigins := os.Getenv("GEEE_ALLOWED_ORIGINS"); envOrigins != "" {
		*allowedOrigins = envOrigins
	}
	if envVerbose := os.Getenv("GEEE_VERBOSE"); envVerbose != "" {
		*verbose = envVerbose == "true" || envVerbose == "1"
	}

	// Initialize logger
	logger := observability.NewJSONLogger(os.Stdout, *verbose)

	logger.Info("server_initializing", map[string]interface{}{
		"port":             *port,
		"host":             *host,
		"cors_enabled":     *enableCORS,
		"allowed_origins":  *allowedOrigins,
		"read_timeout_s":   *readTimeout,
		"write_timeout_s":  *writeTimeout,
		"shutdown_timeout": *shutdownTimeout,
	})

	// Initialize plugin registry
	pluginRegistry := registry.NewDefaultRegistry()

	// Register all plugins
	if err := plugins.RegisterAll(pluginRegistry); err != nil {
		logger.Error("plugin_registration_failed", map[string]interface{}{
			"error": err.Error(),
		})
		fmt.Fprintf(os.Stderr, "Failed to register plugins: %v\n", err)
		os.Exit(1)
	}

	pluginIDs := pluginRegistry.List()
	logger.Info("plugins_registered", map[string]interface{}{
		"count":   len(pluginIDs),
		"plugins": pluginIDs,
	})

	// Initialize metrics collector (Prometheus)
	metrics := observability.NewPrometheusMetricsCollector()

	logger.Info("metrics_initialized", map[string]interface{}{
		"collector": "prometheus",
	})

	// Initialize executor
	exec := executor.NewDefaultExecutor(pluginRegistry, logger, metrics)

	// Parse allowed origins
	origins := []string{"*"}
	if *allowedOrigins != "*" {
		origins = strings.Split(*allowedOrigins, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
	}

	// Create server configuration
	serverConfig := &server.ServerConfig{
		Port:            *port,
		Host:            *host,
		ReadTimeout:     time.Duration(*readTimeout) * time.Second,
		WriteTimeout:    time.Duration(*writeTimeout) * time.Second,
		EnableCORS:      *enableCORS,
		AllowedOrigins:  origins,
		ShutdownTimeout: time.Duration(*shutdownTimeout) * time.Second,
	}

	// Create and start server
	srv := server.NewServer(serverConfig, pluginRegistry, exec, logger, metrics)

	logger.Info("server_ready", map[string]interface{}{
		"address": fmt.Sprintf("http://%s:%d", *host, *port),
	})

	// Start server (blocks until shutdown)
	if err := srv.Start(); err != nil {
		logger.Error("server_error", map[string]interface{}{
			"error": err.Error(),
		})
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
