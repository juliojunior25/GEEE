package httpfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/yourusername/geee/pkg/types"
	"github.com/yourusername/geee/plugins/base"
)

// Config holds the plugin configuration
type Config struct {
	URLField      string `json:"url_field" yaml:"url_field"`
	ResponseField string `json:"response_field" yaml:"response_field"`
	Timeout       int    `json:"timeout" yaml:"timeout"`
	Method        string `json:"method" yaml:"method"`
}

// Response represents an HTTP response
type Response struct {
	Status     int               `json:"status"`
	StatusText string            `json:"status_text"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
}

// Plugin implements HTTP fetching functionality
type Plugin struct {
	*base.BasePlugin
	clientOnce sync.Once
	client     *http.Client
}

// New creates a new http-fetcher plugin instance
func New() *Plugin {
	manifest := types.PluginManifest{
		ID:          "http-fetcher",
		Name:        "HTTP Fetcher",
		Version:     "1.0.0",
		Description: "Fetches content from HTTP/HTTPS URLs",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
		OutputSchema: map[string]interface{}{
			"type": "object",
		},
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url_field": map[string]interface{}{
					"type": "string",
				},
				"response_field": map[string]interface{}{
					"type": "string",
				},
				"timeout": map[string]interface{}{
					"type": "integer",
				},
				"method": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	return &Plugin{
		BasePlugin: base.NewBasePlugin(manifest),
	}
}

func (p *Plugin) getClient() *http.Client {
	p.clientOnce.Do(func() {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.MaxIdleConns = 100
		transport.IdleConnTimeout = 90 * time.Second
		p.client = &http.Client{
			Transport: transport,
		}
	})

	return p.client
}

// Execute performs the HTTP request
func (p *Plugin) Execute(ctx *types.ExecutionContext) (*types.PluginResult, error) {
	// Validate execution context
	if err := p.Validate(ctx); err != nil {
		return nil, err
	}

	// Parse config
	var config Config
	configBytes, err := json.Marshal(ctx.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set defaults
	if config.URLField == "" {
		config.URLField = "url"
	}
	if config.ResponseField == "" {
		config.ResponseField = "response"
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}
	if config.Method == "" {
		config.Method = "GET"
	}

	// Validate method
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "DELETE": true,
		"PATCH": true, "HEAD": true, "OPTIONS": true,
	}
	if !validMethods[config.Method] {
		return nil, fmt.Errorf("invalid HTTP method: %s", config.Method)
	}

	// Get the URL from input
	urlValue, exists := ctx.State[config.URLField]
	if !exists {
		return nil, fmt.Errorf("URL field '%s' not found", config.URLField)
	}

	url, ok := urlValue.(string)
	if !ok {
		return nil, fmt.Errorf("URL field '%s' must be a string, got %T", config.URLField, urlValue)
	}

	client := p.getClient()

	// Create HTTP request with timeout
	reqCtx := ctx.Context
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(reqCtx, time.Duration(config.Timeout)*time.Second)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(reqCtx, config.Method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Extract headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Create response object
	response := Response{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Body:       string(bodyBytes),
		Headers:    headers,
	}

	// Create output map with all original fields
	output := make(map[string]interface{}, len(ctx.State)+1)
	for k, v := range ctx.State {
		output[k] = v
	}
	output[config.ResponseField] = response

	return &types.PluginResult{
		Success: true,
		Data:    output,
	}, nil
}
