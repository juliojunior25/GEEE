# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**GEEE (Generic Extraction & Enrichment Engine)** is a plugin-based data transformation framework written in Go. It's designed for low resource consumption (<50MB idle, <10MB per pipeline) with extreme extensibility through a plugin system. The framework uses AI for planning transformations, not executing code.

**Key Characteristics:**
- Zero memory leaks with guaranteed cleanup
- Plugin-based architecture: extensible without modifying core
- Full observability: traces, metrics, structured logs
- Developer-friendly: Makefile as single interface

## Essential Development Commands

**All development operations MUST use the Makefile.** Never run `go` commands directly.

### Development & Building
```bash
make dev            # Hot reload development mode (uses Air if available)
make build-local    # Build for local OS
make build          # Production build (Linux AMD64)
make run            # Build and execute locally
```

### Testing
```bash
make test           # All tests with coverage
make test-race      # Race detector (required for CI)
make test-coverage  # Detailed coverage report → coverage/coverage.html
make bench          # Run benchmarks
```

### Code Quality
```bash
make fmt            # Format code
make lint           # Run golangci-lint
make vet            # Run go vet
```

### HTTP Server
```bash
make server         # Build and start HTTP server on :8080
make server-dev     # Start server in verbose mode
make server-build   # Build server binary locally
```

### Plugins
```bash
make plugin-new NAME=my-plugin      # Create plugin scaffold
make plugin-test PLUGIN=my-plugin   # Test specific plugin
make plugin-bench PLUGIN=my-plugin  # Benchmark plugin
```

### Quick Example
```bash
# Run the quickstart example
cd examples/quickstart
./run.sh

# Or manually
make build-local
./bin/geee run --config pipeline.yaml --input input.json --output result.json
```

## Architecture Overview

### Core Concepts

**1. Plugin System**
- All plugins implement the `core.Plugin` interface (internal/core/interfaces.go)
- Plugins embed `base.BasePlugin` for automatic cleanup and validation
- Each plugin declares its capabilities via a manifest (types.PluginManifest)
- Manifests include JSON schemas for input/output validation

**2. Execution Flow**
```
Input → Plugin Registry → Execution Plan (DAG) → Executor → Output
```

**3. State Management**
- State is a `map[string]interface{}` passed between plugins
- **Critical**: Executor REPLACES state (doesn't merge) when plugin returns result
- Large binary data stored as file references, not in-memory
- State size limit: 10MB per execution

**4. DAG (Directed Acyclic Graph)**
- Located in `internal/executor/dag.go`
- **Important**: Uses step INDICES (not plugin IDs) to support multiple instances of same plugin
- Tracks completion via both step index and plugin ID for dependency resolution
- Enables parallel execution of independent steps

### Critical Implementation Details

**Executor State Handling (internal/executor/executor.go:103-107)**
```go
// When a plugin returns a result, the entire state is REPLACED
if result.result != nil && result.result.Data != nil {
    state = result.result.Data  // Full replacement, not merge
    finalResult = result.result
}
```

**Metadata Always Set (internal/executor/executor.go:143-149)**
```go
// Metadata is ALWAYS set after execution, even if finalResult exists
finalResult.Metadata = &types.ExecutionMetadata{
    ExecutionID: execID,
    StartTime:   startTime,
    EndTime:     time.Now(),
    DurationMs:  duration.Milliseconds(),
}
```

**DAG Step Tracking (internal/executor/dag.go:12-16)**
```go
type DAG struct {
    stepToIdx         map[*core.ExecutionStep]int // Step pointer → index
    completed         map[int]bool                // Step index → completion
    completedPluginID map[string]bool             // Plugin ID → completion (for deps)
}
```

### Key Interfaces

**Plugin Interface** (internal/core/interfaces.go:12-27)
```go
type Plugin interface {
    Manifest() types.PluginManifest
    Execute(ctx *types.ExecutionContext) (*types.PluginResult, error)
    Cleanup(ctx context.Context) error  // Guaranteed to be called
    Validate(ctx *types.ExecutionContext) error
}
```

**Executor Interface** (internal/core/interfaces.go:51-59)
```go
type Executor interface {
    Execute(ctx context.Context, plan *ExecutionPlan, initialState map[string]interface{}) (*types.PluginResult, error)
    ExecutePlugin(ctx context.Context, pluginID string, state map[string]interface{}, config map[string]interface{}) (*types.PluginResult, error)
}
```

## Plugin Development

### Creating a New Plugin

1. **Generate scaffold:**
```bash
make plugin-new NAME=my-plugin
```

2. **Embed BasePlugin:**
```go
type MyPlugin struct {
    *base.BasePlugin
}

func NewMyPlugin() *MyPlugin {
    manifest := types.PluginManifest{
        ID:          "my-plugin",
        Name:        "My Plugin",
        Version:     "1.0.0",
        Type:        "transformation",
        Optional:    false,
        ParallelSafe: true,
        Timeout:     30,
        MaxMemoryMB: 50,
        InputSchema: map[string]interface{}{ /* ... */ },
        OutputSchema: map[string]interface{}{ /* ... */ },
    }

    return &MyPlugin{
        BasePlugin: base.NewBasePlugin(manifest),
    }
}
```

3. **Implement Execute:**
```go
func (p *MyPlugin) Execute(ctx *types.ExecutionContext) (*types.PluginResult, error) {
    // Register cleanup if needed
    p.RegisterCleanup(func(ctx context.Context) error {
        // cleanup logic
        return nil
    })

    // Process data
    result := make(map[string]interface{})
    // ... transform ctx.State into result ...

    return &types.PluginResult{
        Success: true,
        Data:    result,
    }, nil
}
```

4. **Register plugin** in `plugins/registry.go`:
```go
func RegisterAll(registry core.PluginRegistry) error {
    plugins := []core.Plugin{
        // ... existing plugins ...
        NewMyPlugin(),
    }
    // ...
}
```

### Plugin Best Practices

- **Always** register cleanup functions via `p.RegisterCleanup()`
- **Never** modify `ctx.State` directly - return new data via PluginResult.Data
- Use `p.ExecuteWithTimeout()` for automatic timeout handling
- Respect `ctx.Context` for cancellation
- Keep InputSchema/OutputSchema accurate for validation
- Set `ParallelSafe: true` only if plugin has no shared mutable state

## Configuration Files

### Pipeline YAML Format
```yaml
name: pipeline-name
description: What this pipeline does
timeout: 30              # Global timeout in seconds
stop_on_error: true      # Stop on first error vs continue

plugins:
  - id: plugin-name      # Must match registered plugin ID
    config:              # Plugin-specific config
      key: value
    depends_on:          # Plugin IDs that must complete first
      - other-plugin
    optional: false      # If true, errors won't stop pipeline
    timeout: 10          # Override plugin's default timeout
```

### Example Pipeline
See `examples/quickstart/pipeline.yaml` for a working example with regex extraction and field transformation.

## HTTP Server

### Endpoints

**POST /run** - Execute pipeline
```json
{
  "config": "name: test\nplugins:\n  - id: json-transformer\n    config: {...}",
  "input": {"field": "value"}
}
```

**GET /health** - Health check with plugin status

**GET /ready** - Readiness probe (Kubernetes)

**GET /plugins** - List all registered plugins

**GET /metrics** - Prometheus metrics

### Server Configuration
```bash
# Via flags
./bin/geee-server --port 8080 --host 0.0.0.0 --cors --verbose

# Via environment
export GEEE_PORT=8080
export GEEE_ENABLE_CORS=true
export GEEE_ALLOWED_ORIGINS="http://localhost:3000,https://example.com"
```

## Observability

### Structured Errors
All errors implement `StructuredError` (pkg/errors/errors.go):
```go
type StructuredError struct {
    Code       string  // e.g., "SCHEMA_VALIDATION_FAILED"
    Message    string
    PluginID   string
    FieldPath  string  // JSON path to error location
    Suggestion string  // Actionable fix
}
```

### Structured Logging
JSON format logs (internal/observability/logger.go):
```json
{
  "timestamp": "2025-01-11T10:30:45Z",
  "level": "info",
  "execution_id": "exec-abc123",
  "plugin_id": "http-fetcher",
  "event": "plugin_completed",
  "duration_ms": 1200
}
```

### Prometheus Metrics
Available at `/metrics`:
- `geee_execution_duration_seconds` - Pipeline execution time
- `geee_plugin_execution_duration_seconds` - Per-plugin execution time
- `geee_plugin_errors_total` - Plugin error count
- `geee_active_executions` - Currently running pipelines
- `geee_http_request_duration_seconds` - HTTP request latency
- `geee_http_request_total` - HTTP request count

## Testing

### Test Organization
- Unit tests: `*_test.go` alongside source files
- Integration tests: `test/integration/`
- Fixtures: `test/fixtures/`
- Examples: `examples/quickstart/`

### Running Specific Tests
```bash
# Single package
go test -v ./internal/executor/

# Single test
go test -v ./internal/executor/ -run TestExecutorSinglePlugin

# Plugin tests
make plugin-test PLUGIN=json-transformer
```

### Race Detector (Critical)
Always run race detector before commits:
```bash
make test-race
```

Zero race conditions is a hard requirement.

## Common Patterns

### Parallel Execution
The executor automatically parallelizes independent steps. Mark plugins as `ParallelSafe: true` in manifest to enable.

### Dependency Chains
Use `depends_on` in pipeline YAML to create sequential flows:
```yaml
plugins:
  - id: extract-data
  - id: transform-data
    depends_on:
      - extract-data
  - id: enrich-data
    depends_on:
      - transform-data
```

### Optional Plugins
Set `optional: true` for plugins that shouldn't stop pipeline on error:
```yaml
plugins:
  - id: optional-enrichment
    optional: true
```

### Resource Cleanup
The executor GUARANTEES cleanup is called even on panics. Register cleanup in Execute:
```go
func (p *MyPlugin) Execute(ctx *types.ExecutionContext) (*types.PluginResult, error) {
    file, err := os.Create("/tmp/temp.txt")
    if err != nil {
        return nil, err
    }

    // Register cleanup
    p.RegisterCleanup(func(ctx context.Context) error {
        return os.Remove(file.Name())
    })

    // ... use file ...
}
```

## Directory Structure

```
geee/
├── cmd/
│   ├── geee/           # CLI entry point
│   └── geee-server/    # HTTP server entry point
├── internal/
│   ├── cli/            # CLI commands implementation
│   ├── config/         # YAML config loader
│   ├── core/           # Core interfaces (Plugin, Registry, Executor)
│   ├── executor/       # Execution engine with DAG
│   ├── observability/  # Logging, metrics, tracing
│   ├── registry/       # Plugin registry implementation
│   └── server/         # HTTP server (Chi router)
├── plugins/
│   ├── base/           # BasePlugin with cleanup & validation
│   ├── json-transformer/
│   ├── regex-extractor/
│   ├── http-fetcher/
│   ├── template-renderer/
│   └── registry.go     # Central plugin registration
├── pkg/
│   ├── errors/         # Structured error types
│   └── types/          # Shared types (Manifest, Result, Context)
├── examples/
│   └── quickstart/     # Complete working example
├── test/
│   ├── fixtures/       # Test data
│   └── integration/    # Integration tests
└── configs/            # Example pipeline configs
```

## Troubleshooting

### Plugin Not Found
Check plugin is registered in `plugins/registry.go` and included in `RegisterAll()`.

### State Not Passing Between Plugins
Verify plugin returns `PluginResult` with `Data` field populated. Executor replaces state with this data.

### Nil Pointer on Metadata
Executor always sets metadata. If seeing nil metadata, check you're using latest executor code (fixed in Phase 7).

### Pipeline Hangs
Check for circular dependencies in `depends_on`. Validate all dependency plugin IDs exist.

### Race Condition Detected
Run `make test-race` to identify the issue. Common causes:
- Shared mutable state in plugins
- Missing mutex protection
- Goroutine accessing freed resources

## Project Status

**Version:** 1.0.0
**Status:** Production Ready (All 7 phases complete)

See `IMPLEMENTATION_PLAN.md` for detailed implementation history and `PRP.md` for complete requirements.
