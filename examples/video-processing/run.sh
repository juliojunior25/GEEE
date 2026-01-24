#!/bin/bash
#
# Run the GEEE video analysis pipeline
#
# Usage:
#   ./run.sh [video_url]
#
# Examples:
#   ./run.sh
#   ./run.sh "https://www.youtube.com/shorts/dQw4w9WgXcQ"
#   ./run.sh "https://www.tiktok.com/@user/video/1234567890"
#

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
OUTPUT_FILE="$SCRIPT_DIR/result.json"
PIPELINE_CONFIG="$SCRIPT_DIR/pipeline.yaml"

# Load environment variables if .env exists
if [ -f "$SCRIPT_DIR/.env" ]; then
    source "$SCRIPT_DIR/.env"
fi

# Default video URL (educational AI short)
DEFAULT_VIDEO_URL="https://www.youtube.com/shorts/dQw4w9WgXcQ"
VIDEO_URL="${1:-$DEFAULT_VIDEO_URL}"

echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  GEEE Video Analysis Pipeline${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}"
echo ""

# Step 1: Check prerequisites
echo -e "${BLUE}[1/5]${NC} Checking prerequisites..."

check_command() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}✗ $1 not found${NC}"
        return 1
    else
        echo -e "${GREEN}✓ $1 found${NC}"
        return 0
    fi
}

MISSING_DEPS=0

# Check required tools
if ! check_command yt-dlp; then
    echo -e "${YELLOW}  Install: brew install yt-dlp${NC}"
    MISSING_DEPS=1
fi

if ! check_command ffmpeg; then
    echo -e "${YELLOW}  Install: brew install ffmpeg${NC}"
    MISSING_DEPS=1
fi

if ! check_command whisper; then
    echo -e "${YELLOW}  Install: pip install openai-whisper${NC}"
    MISSING_DEPS=1
fi

# Check LLM server
if curl -s -f http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ LLM server is running${NC}"
else
    echo -e "${YELLOW}⚠ LLM server not detected at localhost:8080${NC}"
    echo -e "${YELLOW}  Start it with: ../../scripts/start-llm.sh${NC}"
    echo ""
    read -p "Continue anyway? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

if [ $MISSING_DEPS -eq 1 ]; then
    echo ""
    echo -e "${RED}Missing required dependencies. Please install them first.${NC}"
    exit 1
fi

echo ""

# Step 2: Check if GEEE is built
echo -e "${BLUE}[2/5]${NC} Checking GEEE binary..."

GEEE_BIN="$PROJECT_ROOT/bin/geee"

if [ ! -f "$GEEE_BIN" ]; then
    echo -e "${YELLOW}GEEE binary not found. Building...${NC}"
    cd "$PROJECT_ROOT"
    make build-local
    cd "$SCRIPT_DIR"
    echo -e "${GREEN}✓ Built successfully${NC}"
else
    echo -e "${GREEN}✓ GEEE binary found${NC}"
fi

echo ""

# Step 3: Prepare pipeline config
echo -e "${BLUE}[3/5]${NC} Preparing pipeline configuration..."

# Create pipeline.yaml with the provided URL
cat > "$PIPELINE_CONFIG" << EOF
name: video-analysis-pipeline
description: Analyze video content from social media platforms
timeout: 300
stop_on_error: true

plugins:
  - id: video-downloader
    config:
      video_url: "$VIDEO_URL"
      output_format: "wav"
      quality: "best"
    timeout: 120

  - id: whisper-transcriber
    config:
      model: "base"
      output_format: "json"
    depends_on:
      - video-downloader
    timeout: 180

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
      temperature: 0.3
      max_tokens: 1000
    depends_on:
      - whisper-transcriber
    timeout: 60

  - id: geocoding-enricher
    config:
      api_key: \${GOOGLE_GEOCODING_API_KEY}
      locations_field: "locations"
    depends_on:
      - llm-analyzer
    timeout: 30
    optional: true
EOF

echo -e "${GREEN}✓ Pipeline configured${NC}"
echo -e "  Video URL: ${BLUE}$VIDEO_URL${NC}"
echo ""

# Step 4: Run the pipeline
echo -e "${BLUE}[4/5]${NC} Running pipeline..."
echo -e "${YELLOW}This may take 1-3 minutes depending on video length...${NC}"
echo ""

START_TIME=$(date +%s)

# Run GEEE
if "$GEEE_BIN" run --config "$PIPELINE_CONFIG" --output "$OUTPUT_FILE"; then
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))

    echo ""
    echo -e "${GREEN}✓ Pipeline completed successfully!${NC}"
    echo -e "  Duration: ${BLUE}${DURATION}s${NC}"
    echo ""
else
    echo ""
    echo -e "${RED}✗ Pipeline failed${NC}"
    exit 1
fi

# Step 5: Display results
echo -e "${BLUE}[5/5]${NC} Results summary..."
echo ""

# Check if jq is available for pretty printing
if command -v jq &> /dev/null; then
    echo -e "${GREEN}Summary:${NC}"
    echo -e "  Video: ${BLUE}$(jq -r '.video_title // "N/A"' "$OUTPUT_FILE")${NC}"
    echo -e "  Duration: ${BLUE}$(jq -r '.video_duration // "N/A"' "$OUTPUT_FILE")${NC}"
    echo -e "  Language: ${BLUE}$(jq -r '.transcript_language // .language // "N/A"' "$OUTPUT_FILE")${NC}"
    echo ""
    echo -e "${GREEN}Analysis:${NC}"
    echo -e "  Topic: ${BLUE}$(jq -r '.main_topic // "N/A"' "$OUTPUT_FILE")${NC}"
    echo -e "  Sentiment: ${BLUE}$(jq -r '.sentiment // "N/A"' "$OUTPUT_FILE")${NC}"
    echo ""
    echo -e "${GREEN}Summary:${NC}"
    jq -r '.summary // "N/A"' "$OUTPUT_FILE" | fold -w 60 -s | sed 's/^/  /'
    echo ""
    echo -e "${GREEN}Key Points:${NC}"
    jq -r '.key_points[]? // empty' "$OUTPUT_FILE" | sed 's/^/  • /'
    echo ""

    # Show locations if geocoded
    LOCATION_COUNT=$(jq '.geocoded_locations | length' "$OUTPUT_FILE" 2>/dev/null || echo "0")
    if [ "$LOCATION_COUNT" -gt 0 ]; then
        echo -e "${GREEN}Locations:${NC}"
        jq -r '.geocoded_locations[]? | "  • \(.formatted_address) (\(.latitude), \(.longitude))"' "$OUTPUT_FILE"
        echo ""
    fi

    echo -e "${YELLOW}Full results saved to:${NC} $OUTPUT_FILE"
    echo -e "${YELLOW}View full JSON:${NC} cat $OUTPUT_FILE | jq ."
else
    echo -e "${GREEN}✓ Results saved to: ${BLUE}$OUTPUT_FILE${NC}"
    echo ""
    echo -e "${YELLOW}Install jq for pretty output:${NC}"
    echo -e "  brew install jq"
    echo ""
    echo -e "${YELLOW}View results:${NC}"
    echo -e "  cat $OUTPUT_FILE"
fi

echo ""
echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}  Done!${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}"
