package firecrawlsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/yourusername/geee/pkg/errors"
	"github.com/yourusername/geee/pkg/types"
	"github.com/yourusername/geee/plugins/base"
)

const (
	defaultAPIURL   = "https://api.firecrawl.dev"
	defaultLocation = "Brazil"
	defaultMaxAge   = 172800000
)

type FirecrawlSearch struct {
	*base.BasePlugin
	clientOnce sync.Once
	client     *http.Client
}

type Config struct {
	APIKey string `json:"api_key"`
	APIURL string `json:"api_url"`
}

type searchRequest struct {
	Query         string                   `json:"query"`
	Sources       []string                 `json:"sources"`
	Categories    []map[string]interface{} `json:"categories"`
	Limit         int                      `json:"limit"`
	Location      string                   `json:"location"`
	ScrapeOptions scrapeOptions            `json:"scrapeOptions"`
}

type scrapeOptions struct {
	OnlyMainContent bool           `json:"onlyMainContent"`
	MaxAge          int            `json:"maxAge"`
	Parsers         []string       `json:"parsers"`
	Formats         []formatOption `json:"formats"`
}

type formatOption struct {
	Type   string                 `json:"type"`
	Schema map[string]interface{} `json:"schema"`
}

type searchResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Web []map[string]interface{} `json:"web"`
	} `json:"data"`
	Warning string `json:"warning"`
	ID      string `json:"id"`
}

func New() *FirecrawlSearch {
	manifest := types.PluginManifest{
		ID:          "firecrawl-search",
		Name:        "Firecrawl Search",
		Version:     "1.0.0",
		Description: "Searches the web via Firecrawl and extracts business info",
		Optional:    false,

		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query for Firecrawl",
				},
			},
			"required": []interface{}{"query"},
		},

		OutputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"company_name": map[string]interface{}{
					"type": "string",
				},
				"company_description": map[string]interface{}{
					"type": "string",
				},
				"opening_hours": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"address": map[string]interface{}{
					"type": "string",
				},
			},
		},

		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"api_key": map[string]interface{}{
					"type":        "string",
					"description": "Firecrawl API key",
				},
				"api_url": map[string]interface{}{
					"type":        "string",
					"description": "Firecrawl API base URL",
				},
			},
			"required": []interface{}{"api_key"},
		},

		Timeout:     60,
		MaxMemoryMB: 50,
	}

	return &FirecrawlSearch{
		BasePlugin: base.NewBasePlugin(manifest),
	}
}

func (p *FirecrawlSearch) getClient() *http.Client {
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

func (p *FirecrawlSearch) Execute(ctx *types.ExecutionContext) (*types.PluginResult, error) {
	if err := p.Validate(ctx); err != nil {
		return nil, err
	}

	var config Config
	if ctx.Config != nil {
		configBytes, _ := json.Marshal(ctx.Config)
		json.Unmarshal(configBytes, &config)
	}

	if config.APIURL == "" {
		config.APIURL = defaultAPIURL
	}

	queryValue, ok := ctx.State["query"].(string)
	if !ok || strings.TrimSpace(queryValue) == "" {
		return nil, errors.NewValidationError(
			"query must be a non-empty string",
			"$.query",
			"Provide a search query in the input state",
		)
	}

	result, warning, err := p.search(ctx, config, queryValue)
	if err != nil {
		return nil, err
	}

	if warning != "" {
		ctx.Logger.Warn("Firecrawl returned warning", map[string]interface{}{
			"warning": warning,
		})
	}

	companyFields, err := extractCompanyFields(ctx, result)
	if err != nil {
		return nil, err
	}

	output := make(map[string]interface{}, len(ctx.State)+len(companyFields))
	for key, value := range ctx.State {
		output[key] = value
	}
	for key, value := range companyFields {
		output[key] = value
	}

	return &types.PluginResult{
		Success: true,
		Data:    output,
	}, nil
}

func (p *FirecrawlSearch) search(ctx *types.ExecutionContext, config Config, query string) (map[string]interface{}, string, error) {
	requestBody := searchRequest{
		Query:      query,
		Sources:    []string{"web"},
		Categories: []map[string]interface{}{},
		Limit:      1,
		Location:   defaultLocation,
		ScrapeOptions: scrapeOptions{
			OnlyMainContent: false,
			MaxAge:          defaultMaxAge,
			Parsers:         []string{"pdf"},
			Formats: []formatOption{
				{
					Type:   "json",
					Schema: defaultJSONSchema(),
				},
			},
		},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := strings.TrimRight(config.APIURL, "/") + "/v2/search"
	req, err := http.NewRequestWithContext(ctx.Context, http.MethodPost, endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	client := p.getClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("firecrawl request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", errors.NewExecutionError(
			"firecrawl-search",
			fmt.Sprintf("Firecrawl API error (status %d)", resp.StatusCode),
			string(bodyBytes),
		)
	}

	var response searchResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, "", fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, response.Warning, errors.NewExecutionError(
			"firecrawl-search",
			"Firecrawl search returned success=false",
			"Check the query and Firecrawl account status",
		)
	}

	if len(response.Data.Web) == 0 {
		return nil, response.Warning, errors.NewExecutionError(
			"firecrawl-search",
			"no search results returned",
			"Try a broader query or adjust the location",
		)
	}

	return response.Data.Web[0], response.Warning, nil
}

func defaultJSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":     "object",
		"required": []interface{}{},
		"properties": map[string]interface{}{
			"company_name": map[string]interface{}{
				"type": "string",
			},
			"company_description": map[string]interface{}{
				"type": "string",
			},
			"opening_hours": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"address": map[string]interface{}{
				"type": "string",
			},
		},
	}
}

func extractCompanyFields(ctx *types.ExecutionContext, result map[string]interface{}) (map[string]interface{}, error) {
	jsonData, ok := extractJSONValue(result["json"])
	if !ok || !hasCompanyFields(jsonData) {
		for _, value := range result {
			candidate, parsed := extractJSONValue(value)
			if parsed && hasCompanyFields(candidate) {
				jsonData = candidate
				ok = true
				break
			}
		}
	}

	if !ok || jsonData == nil {
		return nil, errors.NewExecutionError(
			"firecrawl-search",
			"structured JSON fields not found in Firecrawl response",
			"Ensure Firecrawl returned json data for the requested schema",
		)
	}

	output := make(map[string]interface{})
	missing := []string{}

	if value, exists := jsonData["company_name"]; exists {
		if name, ok := value.(string); ok && strings.TrimSpace(name) != "" {
			output["company_name"] = name
		} else {
			missing = append(missing, "company_name")
		}
	} else {
		missing = append(missing, "company_name")
	}

	if value, exists := jsonData["company_description"]; exists {
		if description, ok := value.(string); ok && strings.TrimSpace(description) != "" {
			output["company_description"] = description
		} else {
			missing = append(missing, "company_description")
		}
	} else {
		missing = append(missing, "company_description")
	}

	if value, exists := jsonData["address"]; exists {
		if address, ok := value.(string); ok && strings.TrimSpace(address) != "" {
			output["address"] = address
		} else {
			missing = append(missing, "address")
		}
	} else {
		missing = append(missing, "address")
	}

	if value, exists := jsonData["opening_hours"]; exists {
		if hours, ok := normalizeStringSlice(value); ok {
			output["opening_hours"] = hours
		} else {
			missing = append(missing, "opening_hours")
		}
	} else {
		missing = append(missing, "opening_hours")
	}

	if len(output) == 0 {
		return nil, errors.NewExecutionError(
			"firecrawl-search",
			"no structured fields extracted",
			"Adjust the query or validate the source page content",
		)
	}

	if len(missing) > 0 {
		ctx.Logger.Warn("Some Firecrawl fields were missing", map[string]interface{}{
			"missing_fields": missing,
		})
	}

	return output, nil
}

func extractJSONValue(value interface{}) (map[string]interface{}, bool) {
	switch typed := value.(type) {
	case map[string]interface{}:
		return typed, true
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return nil, false
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
			return nil, false
		}
		return parsed, true
	default:
		return nil, false
	}
}

func hasCompanyFields(value map[string]interface{}) bool {
	if value == nil {
		return false
	}
	if _, ok := value["company_name"]; ok {
		return true
	}
	if _, ok := value["company_description"]; ok {
		return true
	}
	if _, ok := value["opening_hours"]; ok {
		return true
	}
	if _, ok := value["address"]; ok {
		return true
	}
	return false
}

func normalizeStringSlice(value interface{}) ([]string, bool) {
	switch typed := value.(type) {
	case []string:
		return typed, true
	case []interface{}:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			str, ok := item.(string)
			if !ok {
				return nil, false
			}
			result = append(result, str)
		}
		return result, true
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return nil, false
		}
		return []string{trimmed}, true
	default:
		return nil, false
	}
}
