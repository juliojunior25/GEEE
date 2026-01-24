package geocodingenricher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/yourusername/geee/pkg/errors"
	"github.com/yourusername/geee/pkg/types"
	"github.com/yourusername/geee/plugins/base"
)

// GeocodingEnricher plugin enriches location strings with coordinates using Google Geocoding API
type GeocodingEnricher struct {
	*base.BasePlugin
	clientOnce sync.Once
	client     *http.Client
}

// Config holds the configuration for the geocoding enricher
type Config struct {
	APIKey         string `json:"api_key"`
	LocationsField string `json:"locations_field"`
}

// GeocodedLocation represents a location with coordinates
type GeocodedLocation struct {
	Query            string  `json:"query"`
	FormattedAddress string  `json:"formatted_address"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
}

// GoogleGeocodingResponse represents the Google API response
type GoogleGeocodingResponse struct {
	Results []struct {
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
	} `json:"results"`
	Status string `json:"status"`
}

// New creates a new GeocodingEnricher plugin
func New() *GeocodingEnricher {
	manifest := types.PluginManifest{
		ID:          "geocoding-enricher",
		Name:        "Geocoding Enricher",
		Version:     "1.0.0",
		Description: "Enriches location strings with coordinates using Google Geocoding API",
		Optional:    true, // Don't fail pipeline if API unavailable

		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"locations": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Array of location strings to geocode",
				},
			},
		},

		OutputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"geocoded_locations": map[string]interface{}{
					"type":        "array",
					"description": "Array of enriched location objects",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"query":             map[string]interface{}{"type": "string"},
							"formatted_address": map[string]interface{}{"type": "string"},
							"latitude":          map[string]interface{}{"type": "number"},
							"longitude":         map[string]interface{}{"type": "number"},
						},
					},
				},
			},
		},

		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"api_key": map[string]interface{}{
					"type":        "string",
					"description": "Google Geocoding API key",
				},
				"locations_field": map[string]interface{}{
					"type":        "string",
					"description": "Field name containing locations array",
				},
			},
			"required": []interface{}{"api_key"},
		},

		Timeout:     30, // 30 seconds
		MaxMemoryMB: 20,
	}

	return &GeocodingEnricher{
		BasePlugin: base.NewBasePlugin(manifest),
	}
}

func (p *GeocodingEnricher) getClient() *http.Client {
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

// Execute enriches locations with geocoding data
func (p *GeocodingEnricher) Execute(ctx *types.ExecutionContext) (*types.PluginResult, error) {
	// Validate execution context
	if err := p.Validate(ctx); err != nil {
		return nil, err
	}

	// Parse config
	var config Config
	if ctx.Config != nil {
		configBytes, _ := json.Marshal(ctx.Config)
		json.Unmarshal(configBytes, &config)
	}

	// Set default locations field
	if config.LocationsField == "" {
		config.LocationsField = "locations"
	}

	if config.APIKey == "" {
		return nil, errors.NewValidationError(
			"api_key is required",
			"$.api_key",
			"Provide Google Geocoding API key in config",
		)
	}

	// Extract locations from state
	locationsValue, exists := ctx.State[config.LocationsField]
	if !exists {
		// If locations field doesn't exist and plugin is optional, return unchanged state
		ctx.Logger.Warn("Locations field not found, skipping geocoding", map[string]interface{}{
			"field": config.LocationsField,
		})

		return &types.PluginResult{
			Success: true,
			Data:    ctx.State,
		}, nil
	}

	// Convert to string array
	locations, err := extractLocations(locationsValue)
	if err != nil {
		ctx.Logger.Warn("Invalid locations format, skipping geocoding", map[string]interface{}{
			"error": err.Error(),
		})

		return &types.PluginResult{
			Success: true,
			Data:    ctx.State,
		}, nil
	}

	if len(locations) == 0 {
		ctx.Logger.Info("No locations to geocode", map[string]interface{}{})
		return &types.PluginResult{
			Success: true,
			Data:    ctx.State,
		}, nil
	}

	client := p.getClient()

	// Geocode each location
	geocoded := make([]GeocodedLocation, 0, len(locations))

	requestTimeout := 25 * time.Second // Slightly less than plugin timeout
	for _, location := range locations {
		if location == "" {
			continue
		}

		ctx.Logger.Debug("Geocoding location", map[string]interface{}{
			"query": location,
		})

		reqCtx, cancel := context.WithTimeout(ctx.Context, requestTimeout)
		result, err := geocodeLocation(reqCtx, client, config.APIKey, location)
		cancel()
		if err != nil {
			// Log error but continue (plugin is optional)
			ctx.Logger.Warn("Failed to geocode location", map[string]interface{}{
				"location": location,
				"error":    err.Error(),
			})
			continue
		}

		if result != nil {
			geocoded = append(geocoded, *result)
		}

		// Rate limiting: sleep briefly between requests (Google allows ~50 QPS)
		time.Sleep(100 * time.Millisecond)
	}

	// Preserve state and add geocoded locations
	output := make(map[string]interface{}, len(ctx.State)+1)
	for k, v := range ctx.State {
		output[k] = v
	}

	output["geocoded_locations"] = geocoded

	ctx.Logger.Info("Geocoding completed", map[string]interface{}{
		"execution_id":    ctx.ExecutionID,
		"total_locations": len(locations),
		"geocoded_count":  len(geocoded),
	})

	return &types.PluginResult{
		Success: true,
		Data:    output,
	}, nil
}

// Helper functions

// extractLocations converts various location formats to string array
func extractLocations(value interface{}) ([]string, error) {
	switch v := value.(type) {
	case []interface{}:
		locations := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				locations = append(locations, str)
			}
		}
		return locations, nil

	case []string:
		return v, nil

	case string:
		// Single location as string
		return []string{v}, nil

	default:
		return nil, fmt.Errorf("locations must be array or string, got %T", value)
	}
}

// geocodeLocation calls Google Geocoding API for a single location
func geocodeLocation(ctx context.Context, client *http.Client, apiKey string, query string) (*GeocodedLocation, error) {
	baseURL := "https://maps.googleapis.com/maps/api/geocode/json"

	// Build URL with query params
	params := url.Values{}
	params.Add("address", query)
	params.Add("key", apiKey)

	fullURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var geoResp GoogleGeocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		return nil, err
	}

	if geoResp.Status != "OK" {
		return nil, fmt.Errorf("geocoding failed: %s", geoResp.Status)
	}

	if len(geoResp.Results) == 0 {
		return nil, fmt.Errorf("no results found")
	}

	result := geoResp.Results[0]

	return &GeocodedLocation{
		Query:            query,
		FormattedAddress: result.FormattedAddress,
		Latitude:         result.Geometry.Location.Lat,
		Longitude:        result.Geometry.Location.Lng,
	}, nil
}
