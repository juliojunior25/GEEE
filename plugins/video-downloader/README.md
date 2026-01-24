# Video Downloader Plugin

## Overview
Downloads videos from TikTok, Instagram Reels, and YouTube Shorts using yt-dlp, and optionally extracts audio to WAV format optimized for transcription.

## Version
1.0.0

## Dependencies
- **yt-dlp**: `pip install yt-dlp` or `brew install yt-dlp`
- **ffmpeg**: `brew install ffmpeg` (required for audio extraction)

## Configuration

```yaml
- id: video-downloader
  config:
    video_url: "https://..."  # Can be set here (preferred) or passed in state
    output_format: mp4         # optional: mp4, webm, mkv (default: mp4)
    quality: best              # optional: best, worst, 720p, 480p (default: best)
    extract_audio: true        # optional: extract audio to WAV (default: true)
```

## Input Schema

The plugin accepts `video_url` from either:
1. **Config** (preferred): Set in pipeline YAML
2. **State**: Passed from previous plugin or input

```json
{
  "video_url": "https://..."  // Required if not in config
}
```

## Output Schema

```json
{
  "video_url": "https://...",
  "video_file": "file:///tmp/geee-exec-xxx/video.mp4",
  "audio_file": "file:///tmp/geee-exec-xxx/audio.wav",
  "duration": 45.2,
  "metadata": {
    "title": "Video title",
    "uploader": "username",
    "upload_date": "20260113",
    "platform": "tiktok"
  }
}
```

## Supported Platforms
- TikTok
- Instagram Reels
- YouTube Shorts
- YouTube videos
- Any platform supported by yt-dlp (hundreds of sites)

## Features

### Video Download
- Uses yt-dlp to download videos from multiple platforms
- Automatic format selection based on quality setting
- Extracts metadata (title, uploader, date, platform)
- Returns file:// reference (not loaded into memory)

### Audio Extraction
- Extracts audio to WAV format using ffmpeg
- Optimized settings for Whisper transcription:
  - 16kHz sample rate
  - Mono channel
  - PCM 16-bit encoding
- Optional: can be disabled with `extract_audio: false`

### Resource Management
- Files stored in `/tmp/geee-exec-{execution_id}/`
- Automatic cleanup after pipeline completion
- Guaranteed cleanup even if pipeline is cancelled

## Resource Usage
- **Timeout**: 300s (5 minutes) - can be overridden
- **Memory**: 100MB
- **Disk**: Varies by video size (automatically cleaned up)

## Error Handling

### Missing Dependencies
Returns helpful error if yt-dlp or ffmpeg not installed:
```
yt-dlp is not installed
Suggestion: Install with: pip install yt-dlp or brew install yt-dlp
```

### Download Failures
Captures yt-dlp stderr and includes it in error message for debugging

### Audio Extraction Failures
Logs warning but continues (non-fatal) if audio extraction fails

## Example Usage

### Basic Usage (URL in config)
```yaml
name: download-video
plugins:
  - id: video-downloader
    config:
      video_url: "https://www.tiktok.com/@user/video/123456789"
      quality: best
      extract_audio: true
```

### URL from Previous Plugin
```yaml
name: download-multiple
plugins:
  - id: url-generator
    # ... generates video_url in state

  - id: video-downloader
    config:
      quality: 720p
    depends_on:
      - url-generator
```

### Command Line Test
```bash
# Create test config
cat > test-download.yaml << 'EOF'
name: test
plugins:
  - id: video-downloader
    config:
      video_url: "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
      quality: worst  # Faster for testing
EOF

# Run
geee run --config test-download.yaml --output result.json

# Check output
cat result.json | jq '.data | keys'
# ["audio_file", "duration", "metadata", "video_file", "video_url"]
```

## Cleanup Behavior

The plugin registers cleanup functions in LIFO order:
1. **Last**: Remove entire temp directory
2. **Middle**: Kill ffmpeg process (if running)
3. **First**: Kill yt-dlp process (if running)

This ensures processes are terminated before files are deleted.

## Notes
- Video files are **not** loaded into memory (respects 10MB state limit)
- Audio is extracted at 16kHz mono for optimal Whisper performance
- Metadata is parsed from yt-dlp's `video.info.json` output
- File references use `file://` prefix for consistency with GEEE patterns

## Troubleshooting

### "yt-dlp not found in PATH"
```bash
# Install yt-dlp
pip install yt-dlp
# or
brew install yt-dlp
```

### "ffmpeg is not installed"
```bash
# Install ffmpeg
brew install ffmpeg
# or (Ubuntu/Debian)
sudo apt-get install ffmpeg
```

### "yt-dlp failed" with certificate errors
```bash
# Update certificates
pip install --upgrade certifi
# or update yt-dlp
pip install --upgrade yt-dlp
```

### Download is very slow
- Use `quality: worst` or `quality: 480p` for faster downloads
- Check your internet connection
- Some platforms may rate-limit downloads

## See Also
- [yt-dlp documentation](https://github.com/yt-dlp/yt-dlp)
- [ffmpeg documentation](https://ffmpeg.org/documentation.html)
- [GEEE Plugin Development Guide](../../CLAUDE.md)
