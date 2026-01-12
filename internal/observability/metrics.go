package observability

import (
	"time"

	"github.com/yourusername/geee/pkg/types"
)

// NoOpMetricsCollector is a no-op implementation of MetricsCollector
// This is used when metrics collection is disabled
type NoOpMetricsCollector struct{}

// NewNoOpMetricsCollector creates a new no-op metrics collector
func NewNoOpMetricsCollector() *NoOpMetricsCollector {
	return &NoOpMetricsCollector{}
}

// IncrementCounter is a no-op
func (m *NoOpMetricsCollector) IncrementCounter(name string, labels map[string]string) {
	// No-op
}

// RecordDuration is a no-op
func (m *NoOpMetricsCollector) RecordDuration(name string, duration time.Duration, labels map[string]string) {
	// No-op
}

// SetGauge is a no-op
func (m *NoOpMetricsCollector) SetGauge(name string, value float64, labels map[string]string) {
	// No-op
}

// Ensure NoOpMetricsCollector implements types.MetricsCollector
var _ types.MetricsCollector = (*NoOpMetricsCollector)(nil)
