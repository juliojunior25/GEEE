# LLM Analyzer Plugin

## Overview
Analyzes text using llama.cpp server with **dynamic field extraction**. Users specify which fields to extract in the pipeline config, and the plugin builds a custom prompt for the LLM.

## Version
1.0.0

## Dependencies
- **llama.cpp server**: Running locally with `/completion` endpoint
- Model downloaded (e.g., Llama-2, Mistral, etc.)

## Configuration

```yaml
- id: llm-analyzer
  config:
    api_endpoint: http://localhost:8080/completion  # optional, default shown
    fields_to_extract:                              # required
      - summary
      - resumo
      - entities
      - locations
      - date_mentioned
      - sentiment
    temperature: 0.7        # optional: 0.0-1.0 (default: 0.7)
    max_tokens: 500         # optional: max response tokens (default: 500)
    prompt_template: ""     # optional: custom Go template
```

## Input Schema

Requires `transcript` from previous plugin:

```json
{
  "transcript": "Full text to analyze..."
}
```

## Output Schema

**Dynamic** - adds fields specified in `fields_to_extract`:

```json
{
  "transcript": "Full text...",
  "summary": "Brief summary of content",
  "resumo": "Resumo em português",
  "entities": ["Person1", "Organization2"],
  "locations": ["San Francisco", "New York"],
  "date_mentioned": "January 2026",
  "sentiment": "positive"
}
```

## Features

### Dynamic Field Extraction
The plugin's unique feature - extract **any fields** you specify:

```yaml
fields_to_extract:
  - summary              # Brief summary
  - key_topics           # Main topics discussed
  - people_mentioned     # Names of people
  - companies            # Company names
  - technical_terms      # Technical vocabulary
  - action_items         # Actionable tasks mentioned
  - custom_field_name    # Any field name you want!
```

The LLM will attempt to extract all specified fields and return them as JSON.

### Default Prompt
When no `prompt_template` is provided, the plugin generates a prompt like:

```
Analyze the following transcript and extract the requested information.

Transcript:
[your transcript here]

Extract the following fields and return as JSON:
- summary
- entities
- locations

Return ONLY valid JSON with these exact field names.

JSON output:
```

### Custom Prompt Templates
Use Go templates for full control:

```yaml
prompt_template: |
  Analise este transcript em português.

  Transcript:
  {{.Transcript}}

  Extraia os seguintes campos: {{range .Fields}}{{.}}, {{end}}

  Retorne JSON válido:
```

Template variables:
- `{{.Transcript}}` - The full transcript text
- `{{.Fields}}` - Array of field names to extract

## Resource Usage
- **Timeout**: 60s (1 minute) - can be overridden
- **Memory**: 50MB (minimal, just HTTP calls)
- **Network**: Requires llama.cpp server accessible

## LLM Configuration

### Temperature
Controls randomness:
- `0.0` - Deterministic, factual
- `0.5` - Balanced (recommended)
- `0.7` - Default, creative
- `1.0` - Very creative, less factual

### Max Tokens
Limits response length:
- `300` - Short responses
- `500` - Default, good for most uses
- `800` - Longer, detailed analysis
- `1000+` - Very detailed

## Error Handling

### LLM Server Not Running
```
LLM API request failed: dial tcp [::1]:8080: connect: connection refused
```
**Solution**: Start llama.cpp server first

### Invalid JSON from LLM
If the LLM returns non-JSON text, the plugin:
1. Logs a warning
2. Returns `raw_analysis` field with the full text
3. Does NOT fail the pipeline

### Missing Fields
If the LLM doesn't return all requested fields:
- Logs a warning with missing field names
- Returns available fields
- Does NOT fail the pipeline

## Example Usage

### Basic Analysis
```yaml
name: analyze-transcript
plugins:
  - id: whisper-transcriber
    # ... transcribes audio

  - id: llm-analyzer
    config:
      fields_to_extract:
        - summary
        - sentiment
    depends_on:
      - whisper-transcriber
```

### Multilingual Analysis
```yaml
- id: llm-analyzer
  config:
    fields_to_extract:
      - summary          # English summary
      - resumo           # Portuguese summary
      - resumen          # Spanish summary
    temperature: 0.5
```

### Comprehensive Analysis
```yaml
- id: llm-analyzer
  config:
    fields_to_extract:
      - summary
      - key_topics
      - entities
      - locations
      - dates
      - sentiment
      - action_items
      - technical_terms
    max_tokens: 800      # Longer response needed
    temperature: 0.3     # More factual
```

### Custom Prompt
```yaml
- id: llm-analyzer
  config:
    fields_to_extract:
      - insights
    prompt_template: |
      You are a video content analyzer.

      Transcript:
      {{.Transcript}}

      Provide deep insights about this content.
      Return JSON: {"insights": "your analysis here"}
```

## JSON Extraction

The plugin handles various LLM output formats:

### Clean JSON
```json
{"summary": "..."}
```

### Markdown Wrapped
```markdown
```json
{"summary": "..."}
```
```

### Text with Embedded JSON
```
Here is the analysis: {"summary": "..."}
```

All are automatically extracted and parsed.

## Cleanup Behavior

The plugin:
1. Closes HTTP connections after each request
2. No temp files created
3. Minimal memory footprint

## Notes
- LLM quality depends on model size (larger = better)
- Complex extractions may require larger `max_tokens`
- Field names should be descriptive (e.g., `people_mentioned` not `pm`)
- LLM may hallucinate - verify critical information

## Troubleshooting

### "connection refused"
Start llama.cpp server:
```bash
llama-server -m models/llama-2-7b.gguf --port 8080
```

### "fields_to_extract is required"
Add at least one field:
```yaml
fields_to_extract:
  - summary
```

### Empty or incorrect fields
- Increase `max_tokens` (LLM may be cut off)
- Decrease `temperature` (more focused)
- Simplify field names
- Try a larger model

### Very slow responses
- Use smaller model
- Reduce `max_tokens`
- Reduce number of fields
- Increase timeout if needed

### LLM returns text instead of JSON
- Plugin will log warning and return `raw_analysis`
- Try adjusting prompt_template
- Some models need explicit JSON formatting instructions

## llama.cpp Server Setup

### Quick Start
```bash
# 1. Download llama.cpp
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp
make

# 2. Download a model
mkdir models
wget https://huggingface.co/TheBloke/Llama-2-7B-GGUF/resolve/main/llama-2-7b.Q4_K_M.gguf -O models/llama-2-7b.gguf

# 3. Start server
./llama-server -m models/llama-2-7b.gguf --port 8080 --ctx-size 4096
```

### Model Recommendations
- **Llama-2-7B**: Good balance, fast
- **Mistral-7B**: Better quality, similar speed
- **Llama-2-13B**: Higher quality, slower
- **Llama-2-70B**: Best quality, needs GPU

## See Also
- [llama.cpp GitHub](https://github.com/ggerganov/llama.cpp)
- [GEEE Plugin Development Guide](../../CLAUDE.md)
- [Go Templates Documentation](https://pkg.go.dev/text/template)
