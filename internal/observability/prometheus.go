package observability

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/yourusername/geee/pkg/types"
)

var (
	// Singleton instance
	prometheusInstance *PrometheusMetricsCollector
	prometheusOnce     sync.Once
)

// PrometheusMetricsCollector implements the MetricsCollector interface with Prometheus
type PrometheusMetricsCollector struct {
	// Pipeline metrics
	executionDuration *prometheus.HistogramVec
	executionTotal    *prometheus.CounterVec
	executionErrors   *prometheus.CounterVec
	activeExecutions  prometheus.Gauge

	// Plugin metrics
	pluginDuration *prometheus.HistogramVec
	pluginTotal    *prometheus.CounterVec
	pluginErrors   *prometheus.CounterVec

	// HTTP metrics
	httpRequestDuration *prometheus.HistogramVec
	httpRequestTotal    *prometheus.CounterVec
	httpRequestErrors   *prometheus.CounterVec

	// Resource metrics
	memoryUsage *prometheus.GaugeVec
}

// NewPrometheusMetricsCollector creates a new Prometheus metrics collector (singleton)
func NewPrometheusMetricsCollector() *PrometheusMetricsCollector {
	prometheusOnce.Do(func() {
		collector := &PrometheusMetricsCollector{}

		// Pipeline metrics
		collector.executionDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "geee_execution_duration_seconds",
				Help:    "Duration of pipeline executions in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"pipeline", "status"},
		)
		collector.executionTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geee_execution_total",
				Help: "Total number of pipeline executions",
			},
			[]string{"pipeline", "status"},
		)
		collector.executionErrors = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geee_execution_errors_total",
				Help: "Total number of pipeline execution errors",
			},
			[]string{"pipeline", "error_code"},
		)
		collector.activeExecutions = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "geee_active_executions",
				Help: "Number of currently active pipeline executions",
			},
		)

		// Plugin metrics
		collector.pluginDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "geee_plugin_execution_duration_seconds",
				Help:    "Duration of plugin executions in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"plugin_id", "status"},
		)
		collector.pluginTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geee_plugin_execution_total",
				Help: "Total number of plugin executions",
			},
			[]string{"plugin_id", "status"},
		)
		collector.pluginErrors = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geee_plugin_errors_total",
				Help: "Total number of plugin execution errors",
			},
			[]string{"plugin_id", "error_code"},
		)

		// HTTP metrics
		collector.httpRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "geee_http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		)
		collector.httpRequestTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geee_http_request_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		)
		collector.httpRequestErrors = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geee_http_request_errors_total",
				Help: "Total number of HTTP request errors",
			},
			[]string{"method", "path", "error_type"},
		)

		// Resource metrics
		collector.memoryUsage = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "geee_memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
			[]string{"plugin_id"},
		)

		// Register all metrics
		prometheus.MustRegister(collector.executionDuration)
		prometheus.MustRegister(collector.executionTotal)
		prometheus.MustRegister(collector.executionErrors)
		prometheus.MustRegister(collector.activeExecutions)
		prometheus.MustRegister(collector.pluginDuration)
		prometheus.MustRegister(collector.pluginTotal)
		prometheus.MustRegister(collector.pluginErrors)
		prometheus.MustRegister(collector.httpRequestDuration)
		prometheus.MustRegister(collector.httpRequestTotal)
		prometheus.MustRegister(collector.httpRequestErrors)
		prometheus.MustRegister(collector.memoryUsage)

		prometheusInstance = collector
	})

	return prometheusInstance
}

// IncrementCounter increments a counter metric
func (m *PrometheusMetricsCollector) IncrementCounter(name string, labels map[string]string) {
	switch name {
	case "execution_total":
		m.executionTotal.With(prometheus.Labels(labels)).Inc()
	case "execution_errors":
		m.executionErrors.With(prometheus.Labels(labels)).Inc()
	case "plugin_total":
		m.pluginTotal.With(prometheus.Labels(labels)).Inc()
	case "plugin_errors":
		m.pluginErrors.With(prometheus.Labels(labels)).Inc()
	case "http_request_total":
		m.httpRequestTotal.With(prometheus.Labels(labels)).Inc()
	case "http_request_errors":
		m.httpRequestErrors.With(prometheus.Labels(labels)).Inc()
	}
}

// RecordDuration records a duration metric
func (m *PrometheusMetricsCollector) RecordDuration(name string, duration time.Duration, labels map[string]string) {
	seconds := duration.Seconds()
	switch name {
	case "execution_duration":
		m.executionDuration.With(prometheus.Labels(labels)).Observe(seconds)
	case "plugin_duration":
		m.pluginDuration.With(prometheus.Labels(labels)).Observe(seconds)
	case "http_request_duration":
		m.httpRequestDuration.With(prometheus.Labels(labels)).Observe(seconds)
	}
}

// SetGauge sets a gauge metric
func (m *PrometheusMetricsCollector) SetGauge(name string, value float64, labels map[string]string) {
	switch name {
	case "active_executions":
		m.activeExecutions.Set(value)
	case "memory_usage":
		m.memoryUsage.With(prometheus.Labels(labels)).Set(value)
	}
}

// Ensure PrometheusMetricsCollector implements types.MetricsCollector
var _ types.MetricsCollector = (*PrometheusMetricsCollector)(nil)
