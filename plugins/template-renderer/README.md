# Template Renderer Plugin

## Overview
The Template Renderer plugin renders Go templates with provided data, supporting all standard Go template features.

## Version
1.0.0

## Configuration

```yaml
- id: template-renderer
  config:
    template_field: "template"  # optional, defaults to "template"
    data_field: "data"  # optional, defaults to "data"
    output_field: "result"  # optional, defaults to "result"
```

### Parameters
- `template_field` (optional): Field containing template string (default: "template")
- `data_field` (optional): Field containing template data (default: "data")
- `output_field` (optional): Field name for rendered output (default: "result")

## Example

### Input
```json
{
  "template": "Hello {{.Name}}! You have {{.Count}} messages.",
  "data": {
    "Name": "Alice",
    "Count": 5
  }
}
```

### Configuration
```yaml
- id: template-renderer
  config:
    template_field: "template"
    data_field: "data"
    output_field: "message"
```

### Output
```json
{
  "template": "Hello {{.Name}}! You have {{.Count}} messages.",
  "data": {
    "Name": "Alice",
    "Count": 5
  },
  "message": "Hello Alice! You have 5 messages."
}
```

## Supported Template Features
- Variable interpolation: `{{.FieldName}}`
- Conditionals: `{{if .Condition}}...{{else}}...{{end}}`
- Loops: `{{range .Items}}...{{end}}`
- Functions: `{{.Field | function}}`
- Nested fields: `{{.User.Name}}`
- Comparisons: `{{if eq .Status "active"}}...{{end}}`

## Advanced Example

```json
{
  "template": "{{range .Users}}{{.Name}}: {{.Email}}\\n{{end}}",
  "data": {
    "Users": [
      {"Name": "Alice", "Email": "alice@example.com"},
      {"Name": "Bob", "Email": "bob@example.com"}
    ]
  }
}
```

Output:
```
Alice: alice@example.com
Bob: bob@example.com
```

## Features
- Full Go template syntax support
- Safe template rendering
- Nested data structures
- Conditional rendering
- Iteration support
- Custom field mapping

## Use Cases
- Email template rendering
- Report generation
- Dynamic content generation
- Configuration file generation
- Notification formatting
- Document templating
