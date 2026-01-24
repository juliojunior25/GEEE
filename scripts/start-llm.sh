#!/bin/bash
#
# Start llama.cpp server for GEEE video analysis pipeline
#
# This script automatically downloads and starts a llama.cpp server
# with a recommended model for text analysis tasks.
#
# Usage:
#   ./scripts/start-llm.sh [model_path] [port]
#
# Examples:
#   ./scripts/start-llm.sh                                    # Use default model and port
#   ./scripts/start-llm.sh ./models/llama-2-7b.gguf          # Custom model
#   ./scripts/start-llm.sh ./models/llama-2-7b.gguf 8081     # Custom model and port
#

set -euo pipefail

# Configuration
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEFAULT_PORT=8080
DEFAULT_MODEL_NAME="mistral-7b-instruct-v0.2.Q4_K_M.gguf"
DEFAULT_MODEL_URL="https://huggingface.co/TheBloke/Mistral-7B-Instruct-v0.2-GGUF/resolve/main/mistral-7b-instruct-v0.2.Q4_K_M.gguf"
MODELS_DIR="$ROOT_DIR/models"
LLAMA_CPP_DIR="$ROOT_DIR/llama.cpp"
LLAMA_BUILD_DIR="$LLAMA_CPP_DIR/build"
LLAMA_SERVER_BIN="$LLAMA_BUILD_DIR/bin/llama-server"
DEFAULT_MODEL_PATH="$MODELS_DIR/$DEFAULT_MODEL_NAME"
USE_SYSTEM_LLAMA=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse arguments
MODEL_PATH="${1:-$DEFAULT_MODEL_PATH}"
PORT="${2:-$DEFAULT_PORT}"

if [[ "$MODEL_PATH" != /* ]]; then
    MODEL_PATH="$ROOT_DIR/$MODEL_PATH"
fi

echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  GEEE LLM Server Starter${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}"
echo ""

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if port is available
check_port() {
    if lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 1
    else
        return 0
    fi
}

# Function to download file with progress
download_file() {
    local url=$1
    local output=$2

    echo -e "${YELLOW}Downloading model (this may take a while, ~4GB)...${NC}"

    if command_exists curl; then
        curl -L --progress-bar "$url" -o "$output"
    elif command_exists wget; then
        wget --show-progress "$url" -O "$output"
    else
        echo -e "${RED}Error: Neither curl nor wget found. Please install one.${NC}"
        exit 1
    fi
}

# Step 1: Check if llama.cpp is installed
echo -e "${BLUE}[1/4]${NC} Checking llama.cpp installation..."

if command_exists llama-server; then
    echo -e "${GREEN}✓ Using system llama-server (Homebrew)${NC}"
    LLAMA_SERVER_BIN="$(command -v llama-server)"
    USE_SYSTEM_LLAMA=true
else
    # If llama.cpp already exists but lacks CMake, it may be outdated.
    if [ -d "$LLAMA_CPP_DIR" ] && [ ! -f "$LLAMA_CPP_DIR/CMakeLists.txt" ]; then
        echo -e "${YELLOW}Existing llama.cpp clone is outdated (no CMake).${NC}"
        echo -e "${YELLOW}Please delete $LLAMA_CPP_DIR and re-run this script.${NC}"
        exit 1
    fi

    if [ ! -d "$LLAMA_CPP_DIR" ]; then
        echo -e "${YELLOW}llama.cpp not found. Installing...${NC}"

        # Check for required tools
        if ! command_exists git; then
            echo -e "${RED}Error: git is required but not installed.${NC}"
            exit 1
        fi

        if ! command_exists cmake; then
            echo -e "${RED}Error: cmake is required but not installed.${NC}"
            echo -e "${YELLOW}On macOS: brew install cmake${NC}"
            echo -e "${YELLOW}On Ubuntu: sudo apt-get install cmake${NC}"
            exit 1
        fi

        # Clone llama.cpp
        echo -e "${YELLOW}Cloning llama.cpp repository...${NC}"
        git clone https://github.com/ggml-org/llama.cpp.git "$LLAMA_CPP_DIR"

        # Build llama.cpp server
        echo -e "${YELLOW}Building llama.cpp server...${NC}"
        cmake -S "$LLAMA_CPP_DIR" -B "$LLAMA_BUILD_DIR" -DLLAMA_BUILD_SERVER=ON
        cmake --build "$LLAMA_BUILD_DIR" --config Release

        if [ ! -f "$LLAMA_SERVER_BIN" ]; then
            echo -e "${RED}Error: llama-server binary not found after build.${NC}"
            exit 1
        fi

        echo -e "${GREEN}✓ llama.cpp installed successfully${NC}"
    else
        echo -e "${GREEN}✓ llama.cpp found${NC}"

        if [ ! -f "$LLAMA_SERVER_BIN" ]; then
            if ! command_exists cmake; then
                echo -e "${RED}Error: cmake is required but not installed.${NC}"
                echo -e "${YELLOW}On macOS: brew install cmake${NC}"
                echo -e "${YELLOW}On Ubuntu: sudo apt-get install cmake${NC}"
                exit 1
            fi

            echo -e "${YELLOW}llama-server binary not found. Building...${NC}"
            cmake -S "$LLAMA_CPP_DIR" -B "$LLAMA_BUILD_DIR" -DLLAMA_BUILD_SERVER=ON
            cmake --build "$LLAMA_BUILD_DIR" --config Release

            if [ ! -f "$LLAMA_SERVER_BIN" ]; then
                echo -e "${RED}Error: llama-server binary not found after build.${NC}"
                exit 1
            fi

            echo -e "${GREEN}✓ llama.cpp built successfully${NC}"
        fi
    fi
fi

# Step 2: Check model file
echo -e "${BLUE}[2/4]${NC} Checking model file..."

if [ ! -f "$MODEL_PATH" ]; then
    echo -e "${YELLOW}Model not found at: $MODEL_PATH${NC}"

    # If using default model, offer to download
    if [ "$MODEL_PATH" = "$DEFAULT_MODEL_PATH" ]; then
        echo -e "${YELLOW}Would you like to download the recommended model?${NC}"
        echo -e "  Model: Mistral 7B Instruct v0.2 (Q4_K_M quantization)"
        echo -e "  Size: ~4GB"
        echo -e "  Quality: High (good balance of speed and accuracy)"
        echo ""
        read -p "Download? (y/n): " -n 1 -r
        echo

        if [[ $REPLY =~ ^[Yy]$ ]]; then
            mkdir -p "$MODELS_DIR"
            download_file "$DEFAULT_MODEL_URL" "$MODEL_PATH"
            echo -e "${GREEN}✓ Model downloaded successfully${NC}"
        else
            echo -e "${RED}Error: Model file required. Specify path as first argument.${NC}"
            echo -e "Usage: $0 [model_path] [port]"
            exit 1
        fi
    else
        echo -e "${RED}Error: Model file not found: $MODEL_PATH${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✓ Model found: $MODEL_PATH${NC}"
fi

# Step 3: Check if port is available
echo -e "${BLUE}[3/4]${NC} Checking port availability..."

if ! check_port "$PORT"; then
    echo -e "${YELLOW}Port $PORT is already in use.${NC}"

    # Try to find the process using the port
    PID=$(lsof -ti:$PORT 2>/dev/null || echo "")
    if [ -n "$PID" ]; then
        PROCESS=$(ps -p $PID -o comm= 2>/dev/null || echo "unknown")
        echo -e "Process: ${BLUE}$PROCESS${NC} (PID: $PID)"

        if [[ "$PROCESS" == *"server"* ]] || [[ "$PROCESS" == *"llama"* ]]; then
            echo -e "${GREEN}Looks like llama.cpp server is already running!${NC}"
            echo -e "Endpoint: ${BLUE}http://localhost:$PORT/completion${NC}"
            echo ""
            read -p "Test the connection? (y/n): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                echo -e "${YELLOW}Testing connection...${NC}"
                curl -s -X POST http://localhost:$PORT/completion \
                    -H "Content-Type: application/json" \
                    -d '{"prompt": "Hello", "n_predict": 5}' | grep -q "content" && \
                    echo -e "${GREEN}✓ Server is responding!${NC}" || \
                    echo -e "${RED}✗ Server not responding${NC}"
            fi
            exit 0
        else
            echo -e "${RED}Error: Port is occupied by another process.${NC}"
            echo -e "Options:"
            echo -e "  1. Kill the process: kill $PID"
            echo -e "  2. Use a different port: $0 $MODEL_PATH <different_port>"
            exit 1
        fi
    fi
else
    echo -e "${GREEN}✓ Port $PORT is available${NC}"
fi

# Step 4: Start the server
echo -e "${BLUE}[4/4]${NC} Starting llama.cpp server..."
echo ""
echo -e "${GREEN}Server configuration:${NC}"
echo -e "  Model: ${BLUE}$(basename $MODEL_PATH)${NC}"
echo -e "  Port: ${BLUE}$PORT${NC}"
echo -e "  Endpoint: ${BLUE}http://localhost:$PORT/completion${NC}"
echo -e "  Threads: ${BLUE}$(sysctl -n hw.ncpu 2>/dev/null || nproc 2>/dev/null || echo "auto")${NC}"
echo ""
echo -e "${YELLOW}Starting server (this may take 10-30 seconds)...${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop the server${NC}"
echo ""

# Start the server
LLAMA_SERVER_CWD="$ROOT_DIR"
if [ "$USE_SYSTEM_LLAMA" = false ]; then
    LLAMA_SERVER_CWD="$(dirname "$LLAMA_SERVER_BIN")"
fi

( cd "$LLAMA_SERVER_CWD" && "$LLAMA_SERVER_BIN" \
    --model "$MODEL_PATH" \
    --port "$PORT" \
    --host "127.0.0.1" \
    --ctx-size 2048 \
    --n-gpu-layers 0 \
) 2>&1 | while IFS= read -r line; do
        # Highlight important messages
        if [[ "$line" == *"HTTP server listening"* ]]; then
            echo -e "${GREEN}✓ $line${NC}"
            echo ""
            echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
            echo -e "${GREEN}  Server is ready! You can now run the video analysis pipeline.${NC}"
            echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
            echo ""
            echo -e "Run the pipeline with:"
            echo -e "  ${BLUE}make build-local${NC}"
            echo -e "  ${BLUE}./bin/geee run --config configs/video-analysis-pipeline.yaml --output result.json${NC}"
            echo ""
        elif [[ "$line" == *"error"* ]] || [[ "$line" == *"Error"* ]]; then
            echo -e "${RED}$line${NC}"
        elif [[ "$line" == *"warning"* ]] || [[ "$line" == *"Warning"* ]]; then
            echo -e "${YELLOW}$line${NC}"
        else
            echo "$line"
        fi
    done

# Cleanup on exit
trap "echo -e '\n${YELLOW}Stopping server...${NC}'; exit 0" INT TERM
