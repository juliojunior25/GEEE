package videodownloader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/yourusername/geee/pkg/errors"
	"github.com/yourusername/geee/pkg/types"
	"github.com/yourusername/geee/plugins/base"
)

// VideoDownloader plugin downloads videos from social media using yt-dlp
type VideoDownloader struct {
	*base.BasePlugin
}

// Config holds the configuration for the video downloader
type Config struct {
	VideoURL     string `json:"video_url"`
	OutputFormat string `json:"output_format"`
	Quality      string `json:"quality"`
	ExtractAudio bool   `json:"extract_audio"`
}

// New creates a new VideoDownloader plugin
func New() *VideoDownloader {
	manifest := types.PluginManifest{
		ID:          "video-downloader",
		Name:        "Video Downloader",
		Version:     "1.0.0",
		Description: "Downloads videos from TikTok, Instagram Reels, YouTube Shorts using yt-dlp and extracts audio",
		Optional:    false,

		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"video_url": map[string]interface{}{
					"type":        "string",
					"description": "URL of the video to download",
				},
			},
		},

		OutputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"video_file": map[string]interface{}{
					"type":        "string",
					"description": "file:// reference to downloaded video",
					"pattern":     "^file://",
				},
				"audio_file": map[string]interface{}{
					"type":        "string",
					"description": "file:// reference to extracted audio WAV file",
					"pattern":     "^file://",
				},
				"duration": map[string]interface{}{
					"type":        "number",
					"description": "Video duration in seconds",
				},
				"metadata": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title":       map[string]interface{}{"type": "string"},
						"uploader":    map[string]interface{}{"type": "string"},
						"upload_date": map[string]interface{}{"type": "string"},
						"platform":    map[string]interface{}{"type": "string"},
					},
				},
			},
		},

		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"video_url": map[string]interface{}{
					"type":        "string",
					"description": "URL of the video to download (can be set here instead of input)",
				},
				"output_format": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"mp4", "webm", "mkv"},
					"description": "Video format",
				},
				"quality": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"best", "worst", "720p", "480p"},
					"description": "Video quality preference",
				},
				"extract_audio": map[string]interface{}{
					"type":        "boolean",
					"description": "Extract audio to WAV format for transcription",
				},
			},
		},

		Timeout:     300, // 5 minutes for downloads
		MaxMemoryMB: 100, // Minimal memory since files are on disk
	}

	return &VideoDownloader{
		BasePlugin: base.NewBasePlugin(manifest),
	}
}

// Execute downloads video and optionally extracts audio
func (p *VideoDownloader) Execute(ctx *types.ExecutionContext) (*types.PluginResult, error) {
	// Validate execution context
	if err := p.Validate(ctx); err != nil {
		return nil, err
	}

	// Parse config with defaults
	var config Config
	if ctx.Config != nil {
		configBytes, _ := json.Marshal(ctx.Config)
		json.Unmarshal(configBytes, &config)
	}

	// Set defaults
	if config.OutputFormat == "" {
		config.OutputFormat = "mp4"
	}
	if config.Quality == "" {
		config.Quality = "best"
	}
	// Default extract_audio to true
	if ctx.Config == nil || ctx.Config["extract_audio"] == nil {
		config.ExtractAudio = true
	}

	// Get video URL - prioritize config, fallback to state
	videoURL := config.VideoURL
	if videoURL == "" {
		if urlValue, exists := ctx.State["video_url"]; exists {
			if urlStr, ok := urlValue.(string); ok {
				videoURL = urlStr
			}
		}
	}

	if videoURL == "" {
		return nil, errors.NewValidationError(
			"video_url is required (either in config or state)",
			"$.video_url",
			"Add video_url to plugin config or pass it in state",
		)
	}

	// Check if yt-dlp is installed
	if err := checkYtDlpInstalled(); err != nil {
		return nil, errors.NewExecutionError(
			"video-downloader",
			"yt-dlp is not installed",
			"Install with: pip install yt-dlp or brew install yt-dlp",
		)
	}

	// Create temp directory for this execution
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("geee-exec-%s", ctx.ExecutionID))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Register cleanup to remove temp directory (runs last in LIFO order)
	p.RegisterCleanup(func(cleanupCtx context.Context) error {
		ctx.Logger.Debug("Removing temp directory", map[string]interface{}{
			"path": tempDir,
		})
		return os.RemoveAll(tempDir)
	})

	// Download video
	videoFile, duration, metadata, err := p.downloadVideo(ctx, tempDir, videoURL, config)
	if err != nil {
		return nil, err
	}

	// Extract audio if requested
	var audioFileRef string
	if config.ExtractAudio {
		audioFile, err := p.extractAudio(ctx, tempDir, videoFile)
		if err != nil {
			ctx.Logger.Warn("Audio extraction failed", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			audioFileRef = fmt.Sprintf("file://%s", audioFile)
		}
	}

	// Preserve state and add new fields
	output := make(map[string]interface{})
	for k, v := range ctx.State {
		output[k] = v
	}

	output["video_url"] = videoURL
	output["video_file"] = fmt.Sprintf("file://%s", videoFile)
	if audioFileRef != "" {
		output["audio_file"] = audioFileRef
	}
	output["duration"] = duration
	output["metadata"] = metadata

	ctx.Logger.Info("Video download completed", map[string]interface{}{
		"execution_id": ctx.ExecutionID,
		"video_file":   videoFile,
		"duration":     duration,
		"has_audio":    audioFileRef != "",
	})

	return &types.PluginResult{
		Success: true,
		Data:    output,
	}, nil
}

// downloadVideo downloads the video using yt-dlp
func (p *VideoDownloader) downloadVideo(ctx *types.ExecutionContext, tempDir, videoURL string, config Config) (string, float64, map[string]interface{}, error) {
	// Build yt-dlp command
	outputTemplate := filepath.Join(tempDir, "video.%(ext)s")
	args := []string{
		"--format", config.Quality,
		"--output", outputTemplate,
		"--no-playlist",
		"--write-info-json",
		videoURL,
	}

	// Execute yt-dlp with context for cancellation
	cmd := exec.CommandContext(ctx.Context, "yt-dlp", args...)
	cmd.Dir = tempDir

	// Register cleanup to kill process if cancelled
	p.RegisterCleanup(func(cleanupCtx context.Context) error {
		if cmd.Process != nil {
			return cmd.Process.Kill()
		}
		return nil
	})

	// Capture output for debugging
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	ctx.Logger.Info("Starting video download", map[string]interface{}{
		"url":    videoURL,
		"format": config.Quality,
	})

	if err := cmd.Run(); err != nil {
		ctx.Logger.Error("yt-dlp failed", map[string]interface{}{
			"error":  err.Error(),
			"stdout": stdout.String(),
			"stderr": stderr.String(),
		})
		return "", 0, nil, fmt.Errorf("yt-dlp failed: %w\nStderr: %s", err, stderr.String())
	}

	// Find downloaded video file
	videoFile, err := findVideoFile(tempDir)
	if err != nil {
		return "", 0, nil, fmt.Errorf("failed to find downloaded video: %w", err)
	}

	// Parse metadata from info.json
	metadata, duration := parseMetadata(tempDir)

	return videoFile, duration, metadata, nil
}

// extractAudio extracts audio from video to WAV format using ffmpeg
func (p *VideoDownloader) extractAudio(ctx *types.ExecutionContext, tempDir, videoFile string) (string, error) {
	// Check if ffmpeg is installed
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return "", fmt.Errorf("ffmpeg is not installed (required for audio extraction)")
	}

	audioPath := filepath.Join(tempDir, "audio.wav")

	// Use ffmpeg to extract audio
	// -vn: no video
	// -acodec pcm_s16le: WAV codec
	// -ar 16000: 16kHz sample rate (optimal for Whisper)
	// -ac 1: mono channel
	cmd := exec.CommandContext(ctx.Context, "ffmpeg",
		"-i", videoFile,
		"-vn",
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		audioPath,
	)

	// Register cleanup to kill ffmpeg process
	p.RegisterCleanup(func(cleanupCtx context.Context) error {
		if cmd.Process != nil {
			return cmd.Process.Kill()
		}
		return nil
	})

	ctx.Logger.Info("Extracting audio", map[string]interface{}{
		"video_file": videoFile,
		"audio_file": audioPath,
	})

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w\nStderr: %s", err, stderr.String())
	}

	return audioPath, nil
}

// Helper functions

// checkYtDlpInstalled checks if yt-dlp is available in PATH
func checkYtDlpInstalled() error {
	_, err := exec.LookPath("yt-dlp")
	if err != nil {
		return fmt.Errorf("yt-dlp not found in PATH")
	}
	return nil
}

// findVideoFile finds the downloaded video file in directory
func findVideoFile(dir string) (string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "video.*"))
	if err != nil {
		return "", err
	}

	for _, f := range files {
		ext := filepath.Ext(f)
		// Skip .json and other non-video files
		if ext != ".json" && ext != "" {
			return f, nil
		}
	}

	return "", fmt.Errorf("no video file found in %s", dir)
}

// parseMetadata reads yt-dlp's info.json and extracts metadata
func parseMetadata(dir string) (map[string]interface{}, float64) {
	infoPath := filepath.Join(dir, "video.info.json")
	data, err := os.ReadFile(infoPath)
	if err != nil {
		// Return empty metadata if info.json not found
		return map[string]interface{}{}, 0
	}

	var info map[string]interface{}
	if err := json.Unmarshal(data, &info); err != nil {
		return map[string]interface{}{}, 0
	}

	metadata := map[string]interface{}{
		"title":       getString(info, "title"),
		"uploader":    getString(info, "uploader"),
		"upload_date": getString(info, "upload_date"),
		"platform":    getString(info, "extractor"),
	}

	duration, _ := info["duration"].(float64)

	return metadata, duration
}

// getString safely extracts a string value from a map
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
