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
