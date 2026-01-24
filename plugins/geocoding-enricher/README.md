# Geocoding Enricher Plugin

## Overview
Enriches location strings with geographic coordinates using the Google Geocoding API. This plugin is **optional** - it won't fail your pipeline if the API is unavailable or rate-limited.

## Version
1.0.0

## Dependencies
- **Google Geocoding API**: Requires API key (free tier: $200/month credit)
- **Internet connection**: Required for API calls

## Configuration

```yaml
- id: geocoding-enricher
  config:
    api_key: ${GOOGLE_GEOCODING_API_KEY}  # required
    locations_field: locations             # optional, default: "locations"
  optional: true
```

## Input Schema

Expects a `locations` field (or custom field name) from previous plugin:

```json
{
  "locations": ["San Francisco", "New York", "Tokyo"]
}
```

Or a single location:
```json
{
  "locations": "San Francisco"
}
```

## Output Schema

Adds `geocoded_locations` array to state:

```json
{
  "locations": ["San Francisco", "New York"],
  "geocoded_locations": [
    {
      "query": "San Francisco",
      "formatted_address": "San Francisco, CA, USA",
      "latitude": 37.7749,
      "longitude": -122.4194
    },
    {
      "query": "New York",
      "formatted_address": "New York, NY, USA",
      "latitude": 40.7128,
      "longitude": -74.0060
    }
  ]
}
```

## Features

### Optional Plugin (Graceful Degradation)
This plugin is marked as **optional** (`Optional: true` in manifest):
- If the API key is invalid, pipeline continues
- If the API is rate-limited, pipeline continues
- If locations field is missing, returns state unchanged
- Errors are logged as warnings, not failures

### Flexible Location Input
Accepts multiple formats:
- `[]string` - Array of strings
- `[]interface{}` - Generic array (filters for strings)
- `string` - Single location string

### Rate Limiting
- Automatic 100ms delay between requests
- Prevents hitting Google's rate limits (~50 requests/second max)
- For large batches, consider using a higher delay

### Field Customization
You can specify which field contains locations:

```yaml
config:
  api_key: ${GOOGLE_GEOCODING_API_KEY}
  locations_field: "custom_locations"  # Read from state["custom_locations"]
```

## Resource Usage
- **Timeout**: 30s (configurable)
- **Memory**: 20MB (minimal, just HTTP calls)
- **Network**: ~100-500 bytes per request
- **Rate**: ~10 requests/second (with default 100ms delay)

## Google Geocoding API Setup

### 1. Get API Key

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable "Geocoding API"
4. Navigate to "Credentials" → "Create Credentials" → "API Key"
5. Copy the API key

### 2. Configure Billing (Required)

Google requires billing info even for free tier:
- Free tier: $200/month credit (≈40,000 requests)
- After free tier: $5 per 1,000 requests
- Set up billing alerts to avoid surprises

### 3. Restrict API Key (Recommended)

For security, restrict your API key:
- **Application restrictions**: None or IP addresses
- **API restrictions**: Geocoding API only

### 4. Set Environment Variable

```bash
export GOOGLE_GEOCODING_API_KEY="AIza..."
```

Or use `.env` file:
```bash
GOOGLE_GEOCODING_API_KEY=AIza...
```

## Error Handling

### Missing API Key
```
api_key is required
```
**Solution**: Set `api_key` in config or `GOOGLE_GEOCODING_API_KEY` env var

### Missing Locations Field
```
Locations field not found, skipping geocoding
```
**Behavior**: Returns state unchanged (not an error)

### Invalid Locations Format
```
locations must be array or string, got int
```
**Behavior**: Logs warning, returns state unchanged

### API Error (Rate Limit, Invalid Key, etc.)
```
Failed to geocode location: API error (status 429): OVER_QUERY_LIMIT
```
**Behavior**: Logs warning for failed location, continues with remaining locations

### No Results Found
```
Failed to geocode location: no results found
```
**Behavior**: Skips location, continues with next

## Example Usage

### Basic Usage
```yaml
name: location-analysis
plugins:
  - id: llm-analyzer
    config:
      fields_to_extract:
        - locations
    # ... produces: {"locations": ["Paris", "London"]}

  - id: geocoding-enricher
    config:
      api_key: ${GOOGLE_GEOCODING_API_KEY}
    depends_on:
      - llm-analyzer
```

### Custom Field Name
```yaml
- id: geocoding-enricher
  config:
    api_key: ${GOOGLE_GEOCODING_API_KEY}
    locations_field: "places_mentioned"  # Read from state["places_mentioned"]
```

### With Multiple Sources
```yaml
plugins:
  - id: json-transformer
    config:
      transformations:
        - type: extract
          path: "$.data.cities"
          output_key: "locations"

  - id: geocoding-enricher
    config:
      api_key: ${GOOGLE_GEOCODING_API_KEY}
    depends_on:
      - json-transformer
```

## Output Format Details

### Successful Geocoding
```json
{
  "query": "San Francisco",
  "formatted_address": "San Francisco, CA, USA",
  "latitude": 37.7749295,
  "longitude": -122.4194155
}
```

### Partially Successful
If some locations fail, you'll get results only for successful ones:

Input:
```json
{"locations": ["San Francisco", "InvalidLocationXYZ123", "New York"]}
```

Output:
```json
{
  "geocoded_locations": [
    {"query": "San Francisco", "latitude": 37.7749, ...},
    {"query": "New York", "latitude": 40.7128, ...}
  ]
}
```

Logs will show:
```
WARN Failed to geocode location: InvalidLocationXYZ123 - no results found
```

## Performance Considerations

### Request Speed
- **Default**: ~10 locations/second (100ms delay)
- **Faster**: Remove delay for small batches (risk rate limiting)
- **Slower**: Increase delay for large batches

### Timeout Settings
Default plugin timeout is 30s:
- ~300 locations maximum at 100ms/request
- For larger batches, increase timeout in pipeline:

```yaml
- id: geocoding-enricher
  config:
    api_key: ${GOOGLE_GEOCODING_API_KEY}
  timeout: 120  # 2 minutes for ~1200 locations
```

### Cost Estimation
- Free tier: 40,000 requests/month ($200 credit)
- Paid: $5 per 1,000 requests
- Example: 10 videos/day × 5 locations/video × 30 days = 1,500 requests/month (free)

## Troubleshooting

### "dial tcp: connection refused"
**Cause**: No internet connection or firewall blocking
**Solution**: Check internet connectivity, firewall rules

### "API error (status 403): REQUEST_DENIED"
**Cause**: Invalid API key or Geocoding API not enabled
**Solution**:
1. Verify API key is correct
2. Enable Geocoding API in Google Cloud Console
3. Check API key restrictions

### "API error (status 429): OVER_QUERY_LIMIT"
**Cause**: Rate limit exceeded (50 QPS) or daily quota exceeded
**Solution**:
1. If temporary: Wait 1-2 seconds and retry
2. If persistent: Increase delay between requests
3. Check quota in Google Cloud Console

### "No results found" for valid location
**Cause**: Ambiguous or misspelled location name
**Solution**:
- Be more specific: "Paris, France" instead of "Paris"
- Check spelling
- Use alternative names

### Empty `geocoded_locations` array
**Causes**:
1. All locations failed to geocode
2. Locations field is empty array
3. API errors for all requests

**Debug**:
```yaml
- id: geocoding-enricher
  config:
    api_key: ${GOOGLE_GEOCODING_API_KEY}
  # Add verbose logging
```

Check logs for warnings about failed locations.

## Advanced Usage

### Using Geocoded Data in Next Plugin

```yaml
plugins:
  - id: llm-analyzer
    config:
      fields_to_extract: ["locations"]

  - id: geocoding-enricher
    config:
      api_key: ${GOOGLE_GEOCODING_API_KEY}

  - id: json-transformer
    config:
      transformations:
        - type: extract
          path: "$.geocoded_locations[*].latitude"
          output_key: "latitudes"
        - type: extract
          path: "$.geocoded_locations[*].longitude"
          output_key: "longitudes"
```

### Mapping Locations to Map

```yaml
- id: template-renderer
  config:
    template: |
      Map of mentioned locations:
      {{range .geocoded_locations}}
      - {{.formatted_address}}: https://maps.google.com/?q={{.latitude}},{{.longitude}}
      {{end}}
```

## Cleanup Behavior

The plugin:
1. Closes HTTP connections after completion
2. No temporary files created
3. Minimal memory footprint
4. All cleanup happens automatically

## Notes

- **Accuracy**: Google Geocoding is highly accurate for cities, less so for ambiguous names
- **Caching**: Consider caching results externally if processing same locations repeatedly
- **Alternatives**: For offline geocoding, consider OpenStreetMap Nominatim (requires setup)
- **Privacy**: Location queries are sent to Google (check your privacy requirements)

## See Also
- [Google Geocoding API Documentation](https://developers.google.com/maps/documentation/geocoding)
- [GEEE Plugin Development Guide](../../CLAUDE.md)
- [Optional Plugin Pattern](../../PRP.md#optional-plugins)
