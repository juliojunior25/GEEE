# Regex Extractor Plugin

## Overview
The Regex Extractor plugin extracts data from text using regular expression patterns.

## Version
1.0.0

## Configuration

```yaml
- id: regex-extractor
  config:
    input_field: "text"  # optional, defaults to "text"
    patterns:
      - name: "email"
        pattern: "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}"
      - name: "phone"
        pattern: "\\d{3,}"
```

### Parameters
- `patterns` (required): Array of regex patterns to extract
  - `name`: Field name for extracted value
  - `pattern`: Regular expression pattern
- `input_field` (optional): Field containing text to extract from (default: "text")

## Example

### Input
```json
{
  "text": "Contact us at support@example.com or call 5551234567"
}
```

### Configuration
```yaml
- id: regex-extractor
  config:
    patterns:
      - name: "email"
        pattern: "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}"
      - name: "phone"
        pattern: "\\d{10,}"
```

### Output
```json
{
  "text": "Contact us at support@example.com or call 5551234567",
  "email": "support@example.com",
  "phone": "5551234567"
}
```

## Features
- Multiple pattern extraction in one pass
- Flexible regex pattern support
- Configurable input field
- Preserves original text
- First match extraction

## Use Cases
- Email extraction from text
- Phone number parsing
- URL extraction
- Log parsing
- Data scraping
- Text analysis
