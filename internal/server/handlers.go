package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/geee/internal/config"
	"github.com/yourusername/geee/internal/core"
	"github.com/yourusername/geee/pkg/errors"
	"github.com/yourusername/geee/pkg/types"
)

// Handlers contains all HTTP handlers for the server
type Handlers struct {
	registry core.PluginRegistry
	executor core.Executor
	logger   types.Logger
	ready    bool
	readyAt  time.Time
}

// NewHandlers creates a new Handlers instance
func NewHandlers(registry core.PluginRegistry, executor core.Executor, logger types.Logger) *Handlers {
	return &Handlers{
		registry: registry,
		executor: executor,
		logger:   logger,
		ready:    true,
		readyAt:  time.Now(),
	}
}

// HandleRun handles POST /run requests to execute a pipeline
func (h *Handlers) HandleRun(w http.ResponseWriter, r *http.Request) {
	executionID := uuid.New().String()
	start := time.Now()

	h.logger.Info("run_request_received", map[string]interface{}{
		"execution_id": executionID,
	})

	// Parse request
	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid_request_body", map[string]interface{}{
			"execution_id": executionID,
			"error":        err.Error(),
		})
		h.sendError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", "")
		return
	}

	// Validate request
	if req.Config == "" {
		h.sendError(w, http.StatusBadRequest, "MISSING_CONFIG", "Config field is required", "")
		return
	}

	// Load configuration
	var cfg *config.Config
	var err error

	// Check if config is a file path or inline YAML
	if _, statErr := os.Stat(req.Config); statErr == nil {
		// It's a file path
		cfg, err = config.LoadFromFile(req.Config)
	} else {
		// Assume it's inline YAML
		cfg, err = config.LoadFromBytes([]byte(req.Config))
	}

	if err != nil {
		h.logger.Error("config_load_failed", map[string]interface{}{
			"execution_id": executionID,
			"error":        err.Error(),
		})

		// Check if it's a structured error
		if structErr, ok := err.(*errors.StructuredError); ok {
			h.sendErrorFromStructured(w, http.StatusBadRequest, structErr)
			return
		}

		h.sendError(w, http.StatusBadRequest, "CONFIG_LOAD_FAILED", err.Error(), "Check your configuration file syntax")
		return
	}

	// Convert config to execution plan
	plan := cfg.ToExecutionPlan()

	// Validate plan
	if err := core.ValidateExecutionPlan(plan, h.registry); err != nil {
		h.logger.Error("plan_validation_failed", map[string]interface{}{
			"execution_id": executionID,
			"error":        err.Error(),
		})

		if structErr, ok := err.(*errors.StructuredError); ok {
			h.sendErrorFromStructured(w, http.StatusBadRequest, structErr)
			return
		}

		h.sendError(w, http.StatusBadRequest, "PLAN_VALIDATION_FAILED", err.Error(), "")
		return
	}

	// Execute pipeline
	ctx := context.Background()
	if plan.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(plan.Timeout)*time.Second)
		defer cancel()
	}

	result, execErr := h.executor.Execute(ctx, plan, req.Input)

	duration := time.Since(start)
	durationMS := duration.Milliseconds()

	// Prepare response
	resp := RunResponse{
		ExecutionID: executionID,
		DurationMS:  durationMS,
	}

	if execErr != nil {
		h.logger.Error("execution_failed", map[string]interface{}{
			"execution_id": executionID,
			"duration_ms":  durationMS,
			"error":        execErr.Error(),
		})

		resp.Success = false

		// Check if it's a structured error (from errors package)
		if structErr, ok := execErr.(*errors.StructuredError); ok {
			resp.Error = &ErrorResponse{
				Code:       structErr.Code,
				Message:    structErr.Message,
				FieldPath:  structErr.FieldPath,
				Suggestion: structErr.Suggestion,
				PluginID:   structErr.PluginID,
			}
		} else if structErr, ok := execErr.(*types.StructuredError); ok {
			// Check if it's a structured error (from types package)
			resp.Error = &ErrorResponse{
				Code:       structErr.Code,
				Message:    structErr.Message,
				FieldPath:  structErr.FieldPath,
				Suggestion: structErr.Suggestion,
				PluginID:   structErr.PluginID,
			}
		} else {
			resp.Error = &ErrorResponse{
				Code:    "EXECUTION_FAILED",
				Message: execErr.Error(),
			}
		}

		w.WriteHeader(http.StatusInternalServerError)
	} else {
		h.logger.Info("execution_completed", map[string]interface{}{
			"execution_id": executionID,
			"duration_ms":  durationMS,
		})

		resp.Success = true
		if result != nil {
			resp.Data = result.Data
		}

		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(resp)
}

// HandleHealth handles GET /health requests
func (h *Handlers) HandleHealth(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:  "healthy",
		Plugins: make(map[string]string),
	}

	// Check each plugin's health
	pluginIDs := h.registry.List()
	for _, pluginID := range pluginIDs {
		plugin, err := h.registry.Get(pluginID)
		if err != nil {
			resp.Plugins[pluginID] = "error"
		} else {
			// Plugin exists and can be retrieved
			manifest := plugin.Manifest()
			if manifest.ID != "" {
				resp.Plugins[pluginID] = "healthy"
			} else {
				resp.Plugins[pluginID] = "unhealthy"
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleReady handles GET /ready requests (readiness probe)
func (h *Handlers) HandleReady(w http.ResponseWriter, r *http.Request) {
	resp := ReadyResponse{
		Ready: h.ready,
		Since: h.readyAt,
	}

	if !h.ready {
		resp.Message = "Server is not ready"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		resp.Message = "Server is ready"
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(resp)
}

// HandlePlugins handles GET /plugins requests
func (h *Handlers) HandlePlugins(w http.ResponseWriter, r *http.Request) {
	pluginIDs := h.registry.List()
	plugins := make([]PluginInfo, 0, len(pluginIDs))

	for _, pluginID := range pluginIDs {
		plugin, err := h.registry.Get(pluginID)
		if err != nil {
			h.logger.Warn("plugin_not_found", map[string]interface{}{
				"plugin_id": pluginID,
			})
			continue
		}

		manifest := plugin.Manifest()

		// Extract input/output fields from schemas if available
		inputs := extractFieldsFromSchema(manifest.InputSchema)
		outputs := extractFieldsFromSchema(manifest.OutputSchema)

		info := PluginInfo{
			ID:          manifest.ID,
			Name:        manifest.Name,
			Description: manifest.Description,
			Version:     manifest.Version,
			Inputs:      inputs,
			Outputs:     outputs,
		}

		plugins = append(plugins, info)
	}

	resp := PluginsResponse{
		Plugins: plugins,
		Count:   len(plugins),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// sendError sends a structured error response
func (h *Handlers) sendError(w http.ResponseWriter, status int, code, message, suggestion string) {
	resp := RunResponse{
		Success: false,
		Error: &ErrorResponse{
			Code:       code,
			Message:    message,
			Suggestion: suggestion,
		},
	}

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// sendErrorFromStructured sends an error response from a StructuredError
func (h *Handlers) sendErrorFromStructured(w http.ResponseWriter, status int, err *errors.StructuredError) {
	resp := RunResponse{
		Success: false,
		Error: &ErrorResponse{
			Code:       err.Code,
			Message:    err.Message,
			FieldPath:  err.FieldPath,
			Suggestion: err.Suggestion,
			PluginID:   err.PluginID,
		},
	}

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// extractFieldsFromSchema extracts field names from a JSON schema
func extractFieldsFromSchema(schema map[string]interface{}) []string {
	if schema == nil {
		return []string{}
	}

	// Try to extract properties from the schema
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return []string{}
	}

	fields := make([]string, 0, len(props))
	for field := range props {
		fields = append(fields, field)
	}

	return fields
}
