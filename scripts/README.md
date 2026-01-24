# GEEE Helper Scripts

This directory contains helper scripts for running the GEEE video analysis pipeline.

## Available Scripts

### `start-llm.sh`

Automatically sets up and starts a llama.cpp server for the LLM analyzer plugin.

**Features:**
- ✅ Auto-installs llama.cpp if not present
- ✅ Downloads recommended model (Mistral 7B Instruct)
- ✅ Checks port availability
- ✅ Validates existing server instances
- ✅ Colorized output and progress tracking

**Usage:**

```bash
# Start with defaults (port 8080, auto-download model)
./scripts/start-llm.sh

# Use custom model
./scripts/start-llm.sh ./models/my-model.gguf

# Use custom model and port
./scripts/start-llm.sh ./models/my-model.gguf 8081
```

**First Run:**

On first run, the script will:
1. Clone and build llama.cpp (~5 minutes)
2. Offer to download Mistral 7B Instruct model (~4GB, 10-30 minutes)
3. Start the server on port 8080

**Requirements:**
- **macOS**: Xcode Command Line Tools (`xcode-select --install`)
- **Linux**: `build-essential`, `git` (`sudo apt-get install build-essential git`)
- **Disk space**: ~5GB (llama.cpp + model)
- **RAM**: Minimum 8GB (model loads ~4GB into memory)

**Model Recommendations:**

| Model | Size | Speed | Quality | Use Case |
|-------|------|-------|---------|----------|
| Mistral 7B Instruct Q4_K_M (default) | 4GB | Fast | High | Best balance for video analysis |
| Llama 2 7B Q4_0 | 3.5GB | Very Fast | Medium | Quick testing |
| Llama 2 13B Q4_K_M | 7GB | Medium | Very High | Production with powerful hardware |
| Mistral 7B Q8_0 | 7GB | Slow | Highest | Maximum accuracy |

**Downloading Models:**

Models can be downloaded from Hugging Face:

```bash
# Create models directory
mkdir -p models

# Download Mistral 7B Instruct (recommended)
curl -L "https://huggingface.co/TheBloke/Mistral-7B-Instruct-v0.2-GGUF/resolve/main/mistral-7b-instruct-v0.2.Q4_K_M.gguf" \
  -o models/mistral-7b-instruct-v0.2.Q4_K_M.gguf

# Download Llama 2 7B (smaller, faster)
curl -L "https://huggingface.co/TheBloke/Llama-2-7B-GGUF/resolve/main/llama-2-7b.Q4_0.gguf" \
  -o models/llama-2-7b.Q4_0.gguf
```

Find more models at: https://huggingface.co/models?search=gguf

**Troubleshooting:**

### Port Already in Use

```bash
# Check what's using the port
lsof -i :8080

# Kill the process
kill <PID>

# Or use a different port
./scripts/start-llm.sh ./models/model.gguf 8081
```

### Build Failures

**macOS:**
```bash
# Install Xcode Command Line Tools
xcode-select --install
```

**Linux:**
```bash
# Install build tools
sudo apt-get update
sudo apt-get install build-essential git cmake
```

### Model Not Loading

```bash
# Check file integrity
ls -lh models/

# Verify it's a valid GGUF file
file models/mistral-7b-instruct-v0.2.Q4_K_M.gguf
# Should show: GGUF model file...

# Re-download if corrupted
rm models/mistral-7b-instruct-v0.2.Q4_K_M.gguf
./scripts/start-llm.sh  # Will offer to re-download
```

### Server Not Responding

```bash
# Test the endpoint
curl -X POST http://localhost:8080/completion \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello", "n_predict": 5}'

# Should return JSON with "content" field
```

### Out of Memory

If the model doesn't fit in RAM:

1. **Use a smaller quantization:**
   ```bash
   # Q2_K is smallest (lowest quality)
   ./scripts/start-llm.sh ./models/model.Q2_K.gguf
   ```

2. **Use a smaller model:**
   ```bash
   # 3B parameters instead of 7B
   curl -L "https://huggingface.co/TheBloke/Phi-2-GGUF/resolve/main/phi-2.Q4_K_M.gguf" \
     -o models/phi-2.Q4_K_M.gguf
   ./scripts/start-llm.sh ./models/phi-2.Q4_K_M.gguf
   ```

3. **Close other applications** to free up RAM

**Server Configuration:**

The script starts llama.cpp with these parameters:
- `--ctx-size 2048`: Context window (tokens)
- `--n-gpu-layers 0`: CPU-only (set to higher for GPU acceleration)
- `--host 127.0.0.1`: Localhost only (secure by default)
- `--port 8080`: Default port (configurable)

To customize, edit the `./server` command in `start-llm.sh`.

**GPU Acceleration (Optional):**

For faster inference, enable GPU acceleration:

**macOS (Metal):**
```bash
# Edit start-llm.sh and change:
--n-gpu-layers 0
# to:
--n-gpu-layers 32  # Adjust based on GPU memory
```

**Linux (CUDA):**
```bash
# Rebuild llama.cpp with CUDA
cd llama.cpp
make clean
make server LLAMA_CUBLAS=1
cd ..

# Then run with GPU layers
# Edit start-llm.sh: --n-gpu-layers 32
```

**Testing the Server:**

```bash
# Start the server
./scripts/start-llm.sh

# In another terminal, test it:
curl -X POST http://localhost:8080/completion \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Analyze this video transcript: A person discussing climate change and renewable energy.",
    "temperature": 0.3,
    "n_predict": 200
  }'
```

Expected response:
```json
{
  "content": "This transcript discusses environmental topics...",
  "stop": true,
  "tokens_predicted": 45,
  ...
}
```

## Environment Variables

None required for `start-llm.sh`. The server runs entirely locally.

For the full pipeline, set:
```bash
export GOOGLE_GEOCODING_API_KEY="your-api-key"  # For geocoding-enricher plugin
```

## Production Deployment

For production use, consider:

1. **Run as a service** (systemd, launchd, Docker)
2. **Use a reverse proxy** (nginx) for HTTPS
3. **Enable GPU acceleration** for better performance
4. **Monitor memory usage** and set up alerts
5. **Use a larger model** (13B or 70B parameters) for better accuracy

Example systemd service:

```ini
[Unit]
Description=llama.cpp server for GEEE
After=network.target

[Service]
Type=simple
User=geee
WorkingDirectory=/opt/geee
ExecStart=/opt/geee/llama.cpp/server --model /opt/geee/models/model.gguf --port 8080 --host 127.0.0.1
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## See Also

- [llama.cpp documentation](https://github.com/ggerganov/llama.cpp)
- [GEEE LLM Analyzer Plugin](../plugins/llm-analyzer/README.md)
- [Video Analysis Pipeline Config](../configs/video-analysis-pipeline.yaml)
- [Complete Example](../examples/video-processing/README.md)
