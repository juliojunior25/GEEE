# HTTP Fetcher Plugin

## Overview
The HTTP Fetcher plugin fetches content from HTTP/HTTPS URLs with configurable timeout and methods.

## Version
1.0.0

## Configuration

```yaml
- id: http-fetcher
  config:
    url_field: "url"  # optional, defaults to "url"
    response_field: "response"  # optional, defaults to "response"
    timeout: 30  # optional, defaults to 30 seconds
    method: "GET"  # optional, defaults to "GET"
```

### Parameters
- `url_field` (optional): Field containing the URL to fetch (default: "url")
- `response_field` (optional): Field name for response data (default: "response")
- `timeout` (optional): Request timeout in seconds (default: 30)
- `method` (optional): HTTP method: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS (default: "GET")

## Example

### Input
```json
{
  "url": "https://api.example.com/data"
}
```

### Configuration
```yaml
- id: http-fetcher
  config:
    url_field: "url"
    response_field: "api_response"
    timeout: 10
    method: "GET"
```

### Output
```json
{
  "url": "https://api.example.com/data",
  "api_response": {
    "status": 200,
    "status_text": "200 OK",
    "body": "{\"data\": \"value\"}",
    "headers": {
      "Content-Type": "application/json",
      "Content-Length": "18"
    }
  }
}
```

## Features
- Support for all standard HTTP methods
- Configurable timeout
- Full response capture (status, headers, body)
- Automatic connection cleanup
- Context-aware cancellation

## Use Cases
- API data fetching
- Web scraping
- External service integration
- Health check requests
- Webhook triggers
- Data enrichment from external sources
