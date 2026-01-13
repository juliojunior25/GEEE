package base

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/yourusername/geee/pkg/errors"
)

// ResourceTracker tracks resource usage for a plugin execution
type ResourceTracker struct {
	maxMemoryMB int
	timeoutSec  int

	startTime  time.Time
	mu         sync.RWMutex
	isRunning  bool
	initialMem uint64
}

// NewResourceTracker creates a new resource tracker
func NewResourceTracker(maxMemoryMB, timeoutSec int) *ResourceTracker {
	return &ResourceTracker{
		maxMemoryMB: maxMemoryMB,
		timeoutSec:  timeoutSec,
	}
}

// Start begins resource tracking
func (rt *ResourceTracker) Start() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.isRunning = true
	rt.startTime = time.Now()

	// Capture initial memory
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	rt.initialMem = m.Alloc
}

// Stop ends resource tracking
func (rt *ResourceTracker) Stop() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.isRunning = false
}

// CheckLimits checks if resource limits are exceeded
func (rt *ResourceTracker) CheckLimits() error {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	if !rt.isRunning {
		return nil
	}

	// Check memory limit
	if rt.maxMemoryMB > 0 {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		currentMemMB := (m.Alloc - rt.initialMem) / 1024 / 1024
		if currentMemMB > uint64(rt.maxMemoryMB) {
			return errors.NewResourceLimitError(
				"memory",
				fmt.Sprintf("%d MB", currentMemMB),
				fmt.Sprintf("%d MB", rt.maxMemoryMB),
			)
		}
	}

	// Check time limit
	if rt.timeoutSec > 0 {
		elapsed := time.Since(rt.startTime)
		maxDuration := time.Duration(rt.timeoutSec) * time.Second

		if elapsed > maxDuration {
			return errors.NewResourceLimitError(
				"time",
				elapsed.String(),
				maxDuration.String(),
			)
		}
	}

	return nil
}

// GetElapsedTime returns the time elapsed since Start was called
func (rt *ResourceTracker) GetElapsedTime() time.Duration {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	if !rt.isRunning {
		return 0
	}

	return time.Since(rt.startTime)
}

// GetMemoryUsageMB returns the current memory usage in MB
func (rt *ResourceTracker) GetMemoryUsageMB() uint64 {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return (m.Alloc - rt.initialMem) / 1024 / 1024
}
