# Video Processing Example

Complete working example of GEEE's video analysis pipeline that downloads, transcribes, and analyzes videos from TikTok, Instagram Reels, or YouTube Shorts.

## What This Example Does

This pipeline:
1. **Downloads** a video from a social media URL using `yt-dlp`
2. **Transcribes** the audio using OpenAI Whisper
3. **Analyzes** the transcript with a local LLM (llama.cpp) to extract:
   - Summary
   - Main topics
   - Key points
   - Entities (people, organizations)
   - Locations mentioned
   - Sentiment
   - Hashtags
   - And any custom fields you specify
4. **Enriches** locations with GPS coordinates using Google Geocoding API (optional)

## Prerequisites

### Required Tools

```bash
# 1. yt-dlp (video downloader)
# macOS
brew install yt-dlp

# Linux
sudo curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
sudo chmod a+rx /usr/local/bin/yt-dlp

# 2. ffmpeg (audio extraction)
# macOS
brew install ffmpeg

# Linux
sudo apt-get install ffmpeg

# 3. OpenAI Whisper (transcription)
pip install -U openai-whisper

# 4. llama.cpp (LLM server) - Installed automatically by start-llm.sh
# No manual installation needed!
```

### Optional Tools

```bash
# Google Geocoding API key (for location enrichment)
# Get one at: https://console.cloud.google.com/
# Set in .env file (see .env.example)
```

## Quick Start

### 1. Setup

```bash
# From repository root
cd examples/video-processing

# Copy environment template
cp .env.example .env

# Edit .env and add your Google API key (optional)
nano .env
```

### 2. Start LLM Server

```bash
# This will auto-install llama.cpp and download the model on first run
../../scripts/start-llm.sh

# Wait for: "Server is ready!"
# Leave this terminal running
```

### 3. Run the Pipeline

Open a new terminal:

```bash
cd examples/video-processing

# Run with the example video (educational short about AI)
./run.sh

# Or specify your own video URL
./run.sh "https://www.youtube.com/shorts/YOUR_VIDEO_ID"
```

### 4. View Results

```bash
# Pretty print the JSON output
cat result.json | jq .

# Or just view the file
cat result.json
```

## Example Output

```json
{
  "video_url": "https://www.youtube.com/shorts/dQw4w9WgXcQ",
  "video_title": "Understanding Climate Change in 60 Seconds",
  "video_duration": "0:58",
  "transcript": "Climate change is affecting our planet in unprecedented ways. Scientists have observed rising global temperatures, melting ice caps in Antarctica and Greenland, and increasing extreme weather events. We must take action now through renewable energy adoption, reforestation, and sustainable practices. Every individual can make a difference.",
  "transcript_language": "en",
  "summary": "A brief educational video explaining the impacts of climate change and calling for individual and collective action through renewable energy and sustainable practices.",
  "main_topic": "climate change and environmental action",
  "key_points": [
    "Global temperatures are rising",
    "Ice caps in Antarctica and Greenland are melting",
    "Extreme weather events are increasing",
    "Solutions include renewable energy and reforestation",
    "Individual actions matter"
  ],
  "entities": [
    "Antarctica",
    "Greenland"
  ],
  "locations": [
    "Antarctica",
    "Greenland"
  ],
  "sentiment": "urgent but hopeful",
  "language_style": "educational",
  "call_to_action": "Take action through renewable energy adoption and sustainable practices",
  "hashtags": [
    "#climatechange",
    "#sustainability",
    "#renewableenergy",
    "#environment"
  ],
  "duration_seconds": 58,
  "geocoded_locations": [
    {
      "query": "Antarctica",
      "formatted_address": "Antarctica",
      "latitude": -82.862755,
      "longitude": 135.0
    },
    {
      "query": "Greenland",
      "formatted_address": "Greenland",
      "latitude": 71.706936,
      "longitude": -42.604303
    }
  ],
  "metadata": {
    "execution_id": "exec-20250113-103045",
    "start_time": "2025-01-13T10:30:45Z",
    "end_time": "2025-01-13T10:32:18Z",
    "duration_ms": 93000
  }
}
```

## Configuration

### Pipeline YAML

The `pipeline.yaml` file defines the workflow:

```yaml
plugins:
  - id: video-downloader
    config:
      video_url: "https://..."  # Overridden by run.sh argument
      output_format: "wav"
      quality: "best"

  - id: whisper-transcriber
    config:
      model: "base"             # Options: tiny, base, small, medium, large
      language: "en"            # Or omit for auto-detection
      output_format: "json"

  - id: llm-analyzer
    config:
      llm_endpoint: "http://localhost:8080/completion"
      fields_to_extract:
        - summary
        - main_topic
        - key_points
        - entities
        - locations
        - sentiment
        - language_style
        - call_to_action
        - hashtags
        - duration_seconds

  - id: geocoding-enricher
    config:
      api_key: ${GOOGLE_GEOCODING_API_KEY}
    optional: true  # Won't fail if API unavailable
```

### Customizing Extracted Fields

Edit `pipeline.yaml` to extract different data:

```yaml
- id: llm-analyzer
  config:
    fields_to_extract:
      - summary
      - products_mentioned      # NEW: Products featured
      - questions_asked         # NEW: Questions posed in video
      - music_genre            # NEW: Background music style
      - target_audience        # NEW: Intended viewers
      - emotional_tone         # NEW: Emotional approach
      # Add any fields you want!
```

The LLM will attempt to extract any field you specify.

### Whisper Model Selection

Tradeoff between speed and accuracy:

| Model | Size | Speed | Accuracy | Use Case |
|-------|------|-------|----------|----------|
| `tiny` | 39M | Fastest | Lowest | Quick testing |
| `base` | 74M | Fast | Good | **Default - balanced** |
| `small` | 244M | Medium | Better | Longer videos |
| `medium` | 769M | Slow | High | Accent-heavy audio |
| `large` | 1550M | Slowest | Highest | Production quality |

Edit `pipeline.yaml`:
```yaml
- id: whisper-transcriber
  config:
    model: "small"  # Change from "base" to "small"
```

### Supported Platforms

Tested and working:
- ✅ YouTube Shorts
- ✅ TikTok
- ✅ Instagram Reels
- ✅ YouTube regular videos (but designed for short-form)
- ✅ Twitter/X videos
- ✅ Reddit videos

## Troubleshooting

### "yt-dlp: command not found"

Install yt-dlp:
```bash
# macOS
brew install yt-dlp

# Linux
sudo curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
sudo chmod a+rx /usr/local/bin/yt-dlp
```

### "whisper: command not found"

Install OpenAI Whisper:
```bash
pip install -U openai-whisper

# Or with conda
conda install -c conda-forge openai-whisper
```

### "Connection refused" (LLM server)

Make sure llama.cpp server is running:
```bash
../../scripts/start-llm.sh

# In another terminal, test:
curl http://localhost:8080/health
```

### Video Download Failed

Common issues:
1. **Invalid URL**: Check the URL is correct and publicly accessible
2. **Age-restricted**: yt-dlp can't download age-restricted videos
3. **Private/deleted**: Video must be public
4. **Region-blocked**: Use a VPN if video is geo-restricted

Enable verbose logging:
```bash
yt-dlp --verbose "YOUR_URL"
```

### Transcription Quality Issues

If transcription is inaccurate:
1. **Use a larger Whisper model**: Change from `base` to `medium` or `large`
2. **Specify language**: Add `language: "en"` to whisper-transcriber config
3. **Check audio quality**: Poor audio = poor transcription
4. **Try different audio format**: Change `output_format: "mp3"` in video-downloader

### LLM Extraction Incomplete

If LLM doesn't extract all fields:
1. **Increase max_tokens**: Edit `max_tokens: 1000` to `max_tokens: 2000` in pipeline.yaml
2. **Use custom prompt**: Specify `custom_prompt` in llm-analyzer config
3. **Reduce fields**: Extract fewer fields per run
4. **Use better model**: Download a larger LLM model (13B or 70B parameters)

### Geocoding Not Working

If geocoding fails:
1. **Check API key**: Ensure `GOOGLE_GEOCODING_API_KEY` is set in `.env`
2. **Verify billing**: Google requires billing info even for free tier
3. **Check quota**: Free tier is 40,000 requests/month
4. **It's optional**: Pipeline will succeed even if geocoding fails

### Pipeline Timeout

For longer videos, increase timeout in pipeline.yaml:
```yaml
timeout: 600  # 10 minutes instead of 5

plugins:
  - id: video-downloader
    timeout: 180  # 3 minutes for download
```

## Advanced Usage

### Process Multiple Videos

```bash
# Create a list of URLs
cat > videos.txt << EOF
https://www.youtube.com/shorts/VIDEO1
https://www.youtube.com/shorts/VIDEO2
https://www.youtube.com/shorts/VIDEO3
EOF

# Process each video
while read url; do
  echo "Processing: $url"
  ./run.sh "$url"
  mv result.json "result_$(date +%s).json"
done < videos.txt
```

### Extract Custom Fields

Create a custom pipeline for product reviews:

```yaml
- id: llm-analyzer
  config:
    fields_to_extract:
      - product_name
      - brand
      - price_mentioned
      - pros_listed
      - cons_listed
      - rating_given
      - recommendation
      - affiliate_links
      - discount_codes
```

### Use Different LLM Models

```bash
# Download a different model
curl -L "https://huggingface.co/TheBloke/Llama-2-13B-GGUF/resolve/main/llama-2-13b.Q4_K_M.gguf" \
  -o ../../models/llama-2-13b.Q4_K_M.gguf

# Start server with new model
../../scripts/start-llm.sh ../../models/llama-2-13b.Q4_K_M.gguf
```

### Batch Processing with Progress

```bash
#!/bin/bash
# process-batch.sh

URLS=(
  "https://www.youtube.com/shorts/VIDEO1"
  "https://www.youtube.com/shorts/VIDEO2"
  "https://www.youtube.com/shorts/VIDEO3"
)

TOTAL=${#URLS[@]}
for i in "${!URLS[@]}"; do
  echo "[$((i+1))/$TOTAL] Processing: ${URLS[$i]}"
  ./run.sh "${URLS[$i]}" > "result_$((i+1)).json" 2>&1
  echo "✓ Saved to result_$((i+1)).json"
  sleep 5  # Rate limiting
done

echo "✓ Processed $TOTAL videos"
```

### Custom Prompt Template

Create a custom prompt for specific analysis:

```yaml
- id: llm-analyzer
  config:
    custom_prompt: |
      You are a social media analyst. Analyze this video transcript for brand mentions and marketing tactics.

      Transcript:
      {transcript}

      Extract as JSON:
      {
        "brands_mentioned": ["brand1", "brand2"],
        "marketing_tactics": ["tactic1", "tactic2"],
        "target_demographic": "description",
        "engagement_tactics": ["tactic1", "tactic2"],
        "viral_potential": "low/medium/high",
        "content_type": "educational/entertainment/promotional"
      }
```

### Integrate with Your Application

```python
import subprocess
import json

def analyze_video(url):
    """Analyze a video URL and return structured data."""
    result = subprocess.run(
        ['./run.sh', url],
        capture_output=True,
        text=True,
        cwd='examples/video-processing'
    )

    if result.returncode == 0:
        with open('examples/video-processing/result.json') as f:
            return json.load(f)
    else:
        raise Exception(f"Pipeline failed: {result.stderr}")

# Usage
data = analyze_video("https://www.youtube.com/shorts/ABC123")
print(f"Summary: {data['summary']}")
print(f"Sentiment: {data['sentiment']}")
```

## Performance

Typical processing times (MacBook Pro M1, 16GB RAM):

| Video Length | Whisper Model | Total Time | Breakdown |
|--------------|---------------|------------|-----------|
| 30 seconds   | base          | ~45s       | Download: 5s, Transcribe: 15s, LLM: 20s, Geocoding: 5s |
| 60 seconds   | base          | ~90s       | Download: 8s, Transcribe: 30s, LLM: 45s, Geocoding: 7s |
| 30 seconds   | large         | ~120s      | Download: 5s, Transcribe: 90s, LLM: 20s, Geocoding: 5s |

**Optimization tips:**
- Use `tiny` or `base` Whisper model for faster transcription
- Use smaller LLM model (7B instead of 13B)
- Disable geocoding if not needed (`optional: true`)
- Enable GPU acceleration for Whisper and LLM

## See Also

- [Video Downloader Plugin](../../plugins/video-downloader/README.md)
- [Whisper Transcriber Plugin](../../plugins/whisper-transcriber/README.md)
- [LLM Analyzer Plugin](../../plugins/llm-analyzer/README.md)
- [Geocoding Enricher Plugin](../../plugins/geocoding-enricher/README.md)
- [Complete Pipeline Config](../../configs/video-analysis-pipeline.yaml)
- [Helper Scripts](../../scripts/README.md)

## License

This example is part of the GEEE project. See [LICENSE](../../LICENSE) for details.
