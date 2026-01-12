# JSON Transformer Plugin

## Overview
The JSON Transformer plugin transforms and renames JSON fields based on configured mapping rules.

## Version
1.0.0

## Configuration

```yaml
- id: json-transformer
  config:
    mappings:
      - from: "source_field"
        to: "destination_field"
      - from: "old_name"
        to: "new_name"
```

### Parameters
- `mappings` (required): Array of field mapping rules
  - `from`: Source field name
  - `to`: Destination field name

## Example

### Input
```json
{
  "a": 1,
  "b": "text",
  "c": true
}
```

### Configuration
```yaml
- id: json-transformer
  config:
    mappings:
      - from: "a"
        to: "x"
      - from: "b"
        to: "y"
```

### Output
```json
{
  "x": 1,
  "y": "text",
  "c": true
}
```

## Features
- Rename multiple fields in a single pass
- Preserve fields not included in mappings
- Handle nested structures
- Type-safe field transformations

## Use Cases
- Adapting data formats between systems
- Normalizing field names
- Data migration and transformation
- API response restructuring
