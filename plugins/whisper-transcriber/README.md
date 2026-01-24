# Whisper Transcriber Plugin

## Overview
Transcribes audio files using OpenAI's Whisper model via the CLI. Supports multiple model sizes, language detection, and timestamped segments.

## Version
1.0.0

## Dependencies
- **openai-whisper**: `pip install openai-whisper`
- **ffmpeg**: `brew install ffmpeg` (Whisper dependency)

## Configuration

```yaml
- id: whisper-transcriber
  config:
    model_size: base        # optional: tiny, base, small, medium, large (default: base)
    language: ""            # optional: force language ('en', 'pt', 'es'), leave empty for auto-detect
    output_format: json     # optional: txt, json (default: json)
```

## Input Schema

Requires `audio_file` as a `file://` reference from previous plugin:

```json
{
  "audio_file": "file:///tmp/geee-exec-xxx/audio.wav"
}
```

## Output Schema

```json
{
  "audio_file": "file:///tmp/geee-exec-xxx/audio.wav",
  "transcript": "Full transcription text here...",
  "language_detected": "en",
  "segments": [
    {
      "id": 0,
      "start": 0.0,
      "end": 2.5,
      "text": "First segment text"
    },
    {
      "id": 1,
      "start": 2.5,
      "end": 5.0,
      "text": "Second segment text"
    }
  ]
}
```

## Features

### Model Sizes
- **tiny**: Fastest, least accurate (~39M parameters)
- **base**: Good balance (~74M parameters) - default
- **small**: Better accuracy (~244M parameters)
- **medium**: High accuracy (~769M parameters)
- **large**: Best accuracy (~1550M parameters)

### Language Support
- Auto-detection: Leave `language` empty
- Forced language: Set `language: pt` for Portuguese, `language: en` for English, etc.
- Supports 99 languages

### Output Formats
- **json**: Full output with segments and timestamps
- **txt**: Plain text only (no segments or language info)

### Segments
When using JSON format, get timestamped segments useful for:
- Video subtitles
- Jumping to specific parts
- Analyzing speech patterns

## Resource Usage
- **Timeout**: 120s (2 minutes) - can be overridden
- **Memory**: 2048MB (2GB) - Whisper is memory-intensive
- **CPU**: Uses all available cores

## Performance Guidelines

### Model Selection
- **Quick testing**: Use `tiny` (fast, ~1-2x realtime)
- **Production**: Use `base` or `small` (balanced)
- **High accuracy**: Use `medium` or `large` (slower, 5-10x realtime)

### Language Forcing
Forcing language can speed up transcription by 10-20%:
```yaml
language: pt  # Skip language detection
```

## Error Handling

### Missing Whisper
Returns helpful error:
```
OpenAI Whisper is not installed
Suggestion: Install with: pip install openai-whisper
```

### Invalid File Reference
Validates `file://` prefix:
```
audio_file must be a file:// URI, got: /tmp/audio.wav
Suggestion: Audio file must use file:// prefix
```

### File Not Found
Checks file exists before running Whisper:
```
audio file not found: /tmp/geee-exec-xxx/audio.wav
```

## Example Usage

### Basic Usage
```yaml
name: transcribe-audio
plugins:
  - id: video-downloader
    config:
      video_url: "https://..."
      extract_audio: true

  - id: whisper-transcriber
    config:
      model_size: base
    depends_on:
      - video-downloader
```

### Force Portuguese
```yaml
- id: whisper-transcriber
  config:
    model_size: small
    language: pt  # Force Portuguese
```

### Text-Only Output
```yaml
- id: whisper-transcriber
  config:
    model_size: tiny
    output_format: txt  # No segments, just text
```

### Command Line Test
```bash
# Create test pipeline
cat > test-whisper.yaml << 'EOF'
name: test
plugins:
  - id: whisper-transcriber
    config:
      model_size: tiny  # Fastest for testing
      output_format: json
EOF

# Create test input (assuming you have an audio file)
echo '{"audio_file": "file:///path/to/audio.wav"}' > input.json

# Run
geee run --config test-whisper.yaml --input input.json --output result.json

# Check output
cat result.json | jq '.data.transcript'
```

## Cleanup Behavior

The plugin registers cleanup functions for:
1. **Whisper process**: Killed if cancelled
2. **Output directory**: Removed after transcription

Whisper creates output files in `/tmp/geee-whisper-{execution_id}/` which are automatically cleaned up.

## Notes
- Audio file is read from disk (not loaded into memory)
- Best results with 16kHz mono WAV files
- Longer audio = more time and memory
- Supports any audio format that ffmpeg can decode

## Troubleshooting

### "whisper not found in PATH"
```bash
# Install Whisper
pip install openai-whisper

# Verify installation
whisper --help
```

### "failed to read whisper output"
- Check that output directory is writable
- Verify disk space available
- Check whisper stderr in logs

### Very slow transcription
- Use smaller model: `tiny` or `base`
- Force language to skip detection
- Ensure enough RAM available (close other applications)

### Out of memory
- Use smaller model
- Reduce audio length (split into chunks)
- Increase MaxMemoryMB in manifest if needed

## Model Download
First run downloads the selected model (~40MB-3GB depending on size):
```bash
# Pre-download models
whisper --model tiny --download_root ~/.cache/whisper
whisper --model base --download_root ~/.cache/whisper
```

## See Also
- [OpenAI Whisper GitHub](https://github.com/openai/whisper)
- [Whisper Model Card](https://github.com/openai/whisper/blob/main/model-card.md)
- [GEEE Plugin Development Guide](../../CLAUDE.md)
