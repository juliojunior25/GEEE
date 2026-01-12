package server

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/yourusername/geee/pkg/types"
)

// LoggingMiddleware logs all HTTP requests
func LoggingMiddleware(logger types.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Log request
			logger.Info("http_request", map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"remote": r.RemoteAddr,
			})

			// Execute next handler
			next.ServeHTTP(ww, r)

			// Log response
			duration := time.Since(start)
			logger.Info("http_response", map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      ww.Status(),
				"duration_ms": duration.Milliseconds(),
				"bytes":       ww.BytesWritten(),
			})
		})
	}
}

// RecoveryMiddleware recovers from panics and returns 500
func RecoveryMiddleware(logger types.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					// Log the panic
					logger.Error("panic_recovered", map[string]interface{}{
						"error":      fmt.Sprintf("%v", rvr),
						"stack":      string(debug.Stack()),
						"method":     r.Method,
						"path":       r.URL.Path,
						"remote":     r.RemoteAddr,
					})

					// Return 500 error
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"success":false,"error":{"code":"INTERNAL_ERROR","message":"Internal server error"}}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// ContentTypeMiddleware ensures JSON content type for API responses
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// MetricsMiddleware collects HTTP metrics
func MetricsMiddleware(metrics types.MetricsCollector) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Execute next handler
			next.ServeHTTP(ww, r)

			// Record metrics
			duration := time.Since(start)
			status := fmt.Sprintf("%d", ww.Status())

			metrics.RecordDuration("http_request_duration", duration, map[string]string{
				"method": r.Method,
				"path":   r.URL.Path,
				"status": status,
			})

			metrics.IncrementCounter("http_request_total", map[string]string{
				"method": r.Method,
				"path":   r.URL.Path,
				"status": status,
			})

			// Record errors for 5xx status codes
			if ww.Status() >= 500 {
				metrics.IncrementCounter("http_request_errors", map[string]string{
					"method":     r.Method,
					"path":       r.URL.Path,
					"error_type": "server_error",
				})
			} else if ww.Status() >= 400 {
				metrics.IncrementCounter("http_request_errors", map[string]string{
					"method":     r.Method,
					"path":       r.URL.Path,
					"error_type": "client_error",
				})
			}
		})
	}
}
