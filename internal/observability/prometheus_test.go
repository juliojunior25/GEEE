package observability

import (
	"testing"
	"time"
)

func TestNewPrometheusMetricsCollector(t *testing.T) {
	collector := NewPrometheusMetricsCollector()

	if collector == nil {
		t.Fatal("Expected non-nil collector")
	}

	if collector.executionDuration == nil {
		t.Error("Expected executionDuration to be initialized")
	}

	if collector.pluginDuration == nil {
		t.Error("Expected pluginDuration to be initialized")
	}

	if collector.httpRequestDuration == nil {
		t.Error("Expected httpRequestDuration to be initialized")
	}
}

func TestIncrementCounter(t *testing.T) {
	collector := NewPrometheusMetricsCollector()

	tests := []struct {
		name   string
		metric string
		labels map[string]string
	}{
		{"execution_total", "execution_total", map[string]string{"pipeline": "test", "status": "success"}},
		{"execution_errors", "execution_errors", map[string]string{"pipeline": "test", "error_code": "TEST_ERROR"}},
		{"plugin_total", "plugin_total", map[string]string{"plugin_id": "test-plugin", "status": "success"}},
		{"plugin_errors", "plugin_errors", map[string]string{"plugin_id": "test-plugin", "error_code": "TEST_ERROR"}},
		{"http_request_total", "http_request_total", map[string]string{"method": "GET", "path": "/test", "status": "200"}},
		{"http_request_errors", "http_request_errors", map[string]string{"method": "GET", "path": "/test", "error_type": "timeout"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			collector.IncrementCounter(tt.metric, tt.labels)
		})
	}
}

func TestRecordDuration(t *testing.T) {
	collector := NewPrometheusMetricsCollector()

	tests := []struct {
		name     string
		metric   string
		duration time.Duration
		labels   map[string]string
	}{
		{"execution_duration", "execution_duration", 100 * time.Millisecond, map[string]string{"pipeline": "test", "status": "success"}},
		{"plugin_duration", "plugin_duration", 50 * time.Millisecond, map[string]string{"plugin_id": "test-plugin", "status": "success"}},
		{"http_request_duration", "http_request_duration", 10 * time.Millisecond, map[string]string{"method": "GET", "path": "/test", "status": "200"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			collector.RecordDuration(tt.metric, tt.duration, tt.labels)
		})
	}
}

func TestSetGauge(t *testing.T) {
	collector := NewPrometheusMetricsCollector()

	tests := []struct {
		name   string
		metric string
		value  float64
		labels map[string]string
	}{
		{"active_executions", "active_executions", 5, map[string]string{}},
		{"memory_usage", "memory_usage", 1024, map[string]string{"plugin_id": "test-plugin"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			collector.SetGauge(tt.metric, tt.value, tt.labels)
		})
	}
}

func TestPrometheusMetricsCollector_MultipleOperations(t *testing.T) {
	collector := NewPrometheusMetricsCollector()

	// Simulate a complete execution flow using interface methods
	collector.SetGauge("active_executions", 1, map[string]string{})
	collector.RecordDuration("execution_duration", 100*time.Millisecond, map[string]string{
		"pipeline": "test-pipeline",
		"status":   "success",
	})
	collector.IncrementCounter("execution_total", map[string]string{
		"pipeline": "test-pipeline",
		"status":   "success",
	})
	collector.RecordDuration("plugin_duration", 30*time.Millisecond, map[string]string{
		"plugin_id": "plugin1",
		"status":    "success",
	})
	collector.RecordDuration("plugin_duration", 40*time.Millisecond, map[string]string{
		"plugin_id": "plugin2",
		"status":    "success",
	})
	collector.SetGauge("active_executions", 0, map[string]string{})

	// Should not panic
}

func TestPrometheusMetricsCollector_ConcurrentAccess(t *testing.T) {
	collector := NewPrometheusMetricsCollector()

	// Test concurrent access (Prometheus metrics are thread-safe)
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			collector.SetGauge("active_executions", 1, map[string]string{})
			collector.RecordDuration("execution_duration", 10*time.Millisecond, map[string]string{
				"pipeline": "test",
				"status":   "success",
			})
			collector.IncrementCounter("execution_total", map[string]string{
				"pipeline": "test",
				"status":   "success",
			})
			collector.SetGauge("active_executions", 0, map[string]string{})
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
