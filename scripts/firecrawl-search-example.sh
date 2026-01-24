#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
ROOT_DIR=$(cd "$SCRIPT_DIR/.." && pwd)

FIRECRAWL_API_KEY="${FIRECRAWL_API_KEY:-fc-COLE_SEU_TOKEN_AQUI}"
QUERY="${1:-Bar nacional Curitiba}"

if [[ "$FIRECRAWL_API_KEY" == "fc-COLE_SEU_TOKEN_AQUI" ]]; then
  echo "Defina FIRECRAWL_API_KEY com seu token antes de rodar."
  exit 1
fi

cd "$ROOT_DIR"

if [[ ! -f "./bin/geee" ]]; then
  make build-local
fi

CONFIG_PATH="$(mktemp -t geee-firecrawl-config.XXXXXX.yaml)"
INPUT_PATH="$(mktemp -t geee-firecrawl-input.XXXXXX.json)"
OUTPUT_PATH="$(mktemp -t geee-firecrawl-output.XXXXXX.json)"

cleanup() {
  rm -f "$CONFIG_PATH" "$INPUT_PATH" "$OUTPUT_PATH"
}
trap cleanup EXIT

cat > "$CONFIG_PATH" <<EOF
name: firecrawl-search-example
description: Firecrawl search example
timeout: 90
stop_on_error: true
plugins:
  - id: firecrawl-search
    config:
      api_key: "$FIRECRAWL_API_KEY"
    timeout: 60
    optional: false
EOF

cat > "$INPUT_PATH" <<EOF
{"query": "$QUERY"}
EOF

./bin/geee run --config "$CONFIG_PATH" --input "$INPUT_PATH" --output "$OUTPUT_PATH"

cat "$OUTPUT_PATH"
