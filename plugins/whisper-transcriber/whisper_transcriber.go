package whispertranscriber

import (
	"bytes"
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yourusername/geee/pkg/errors"
	"github.com/yourusername/geee/pkg/types"
	"github.com/yourusername/geee/plugins/base"
)

// WhisperTranscriber plugin transcribes audio using OpenAI Whisper CLI
type WhisperTranscriber struct {
	*base.BasePlugin
}

// Config holds the configuration for the whisper transcriber
type Config struct {
	ModelSize    string `json:"model_size"`
	Language     string `json:"language"`
	OutputFormat string `json:"output_format"`
}

// WhisperOutput represents the JSON output from Whisper
type WhisperOutput struct {
	Text     string           `json:"text"`
	Segments []WhisperSegment `json:"segments"`
	Language string           `json:"language"`
}

// WhisperSegment represents a timestamped segment
type WhisperSegment struct {
	ID    int     `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// New creates a new WhisperTranscriber plugin
func New() *WhisperTranscriber {
	manifest := types.PluginManifest{
		ID:          "whisper-transcriber",
		Name:        "Whisper Transcriber",
		Version:     "1.0.0",
		Description: "Transcribes audio files using OpenAI Whisper CLI",
		Optional:    false,

		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"audio_file": map[string]interface{}{
					"type":        "string",
					"description": "file:// reference to audio file",
					"pattern":     "^file://",
				},
			},
			"required": []interface{}{"audio_file"},
		},

		OutputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"transcript": map[string]interface{}{
					"type":        "string",
					"description": "Full transcription text",
				},
				"language_detected": map[string]interface{}{
					"type":        "string",
					"description": "Detected language code (e.g., 'en', 'pt')",
				},
				"segments": map[string]interface{}{
					"type":        "array",
					"description": "Timestamped segments",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"id":    map[string]interface{}{"type": "integer"},
							"start": map[string]interface{}{"type": "number"},
							"end":   map[string]interface{}{"type": "number"},
							"text":  map[string]interface{}{"type": "string"},
						},
					},
				},
			},
		},

		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"model_size": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"tiny", "base", "small", "medium", "large"},
					"description": "Whisper model size (larger = more accurate, slower)",
				},
				"language": map[string]interface{}{
					"type":        "string",
					"description": "Force specific language (e.g., 'en', 'pt', 'es'). Leave empty for auto-detect",
				},
				"output_format": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"txt", "json"},
					"description": "Output format from Whisper",
				},
			},
		},

		Timeout:     120,  // 2 minutes
		MaxMemoryMB: 2048, // Whisper can use significant memory
	}

	return &WhisperTranscriber{
		BasePlugin: base.NewBasePlugin(manifest),
	}
}

// Execute transcribes the audio file
func (p *WhisperTranscriber) Execute(ctx *types.ExecutionContext) (*types.PluginResult, error) {
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
	if config.ModelSize == "" {
		config.ModelSize = "base"
	}
	if config.OutputFormat == "" {
		config.OutputFormat = "json"
	}

	// Check if whisper is installed
	if err := checkWhisperInstalled(); err != nil {
		return nil, errors.NewExecutionError(
			"whisper-transcriber",
			"OpenAI Whisper is not installed",
			"Install with: pip install openai-whisper",
		)
	}

	// Extract and validate audio_file reference
	audioFileRef, ok := ctx.State["audio_file"].(string)
	if !ok {
		return nil, errors.NewValidationError(
			"audio_file must be a string",
			"$.audio_file",
			"Ensure audio_file is a string with file:// prefix",
		)
	}

	if !strings.HasPrefix(audioFileRef, "file://") {
		return nil, errors.NewValidationError(
			fmt.Sprintf("audio_file must be a file:// URI, got: %s", audioFileRef),
			"$.audio_file",
			"Audio file must use file:// prefix (e.g., file:///tmp/audio.wav)",
		)
	}

	audioPath := strings.TrimPrefix(audioFileRef, "file://")

	// Verify audio file exists
	if _, err := os.Stat(audioPath); err != nil {
		return nil, fmt.Errorf("audio file not found: %s", audioPath)
	}

	// Create temp output directory
	outputDir := filepath.Join(os.TempDir(), fmt.Sprintf("geee-whisper-%s", ctx.ExecutionID))
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output dir: %w", err)
	}

	// Register cleanup to remove output directory
	p.RegisterCleanup(func(cleanupCtx context.Context) error {
		ctx.Logger.Debug("Removing whisper output directory", map[string]interface{}{
			"path": outputDir,
		})
		return os.RemoveAll(outputDir)
	})

	// Transcribe audio
	transcript, languageDetected, segments, err := p.transcribeAudio(ctx, audioPath, outputDir, config)
	if err != nil {
		return nil, err
	}

	// Preserve state and add new fields
	output := make(map[string]interface{})
	for k, v := range ctx.State {
		output[k] = v
	}

	output["transcript"] = strings.TrimSpace(transcript)
	if languageDetected != "" {
		output["language_detected"] = languageDetected
	}
	if len(segments) > 0 {
		output["segments"] = segments
	}

	ctx.Logger.Info("Transcription completed", map[string]interface{}{
		"execution_id": ctx.ExecutionID,
		"text_length":  len(transcript),
		"language":     languageDetected,
		"segments":     len(segments),
	})

	return &types.PluginResult{
		Success: true,
		Data:    output,
	}, nil
}

// transcribeAudio runs whisper CLI and parses output
func (p *WhisperTranscriber) transcribeAudio(ctx *types.ExecutionContext, audioPath, outputDir string, config Config) (string, string, []interface{}, error) {
	// Build whisper command
	args := []string{
		audioPath,
		"--model", config.ModelSize,
		"--output_format", config.OutputFormat,
		"--output_dir", outputDir,
	}

	if config.Language != "" {
		args = append(args, "--language", config.Language)
	}

	// Execute whisper with context for cancellation
	cmd := exec.CommandContext(ctx.Context, "whisper", args...)

	// Register cleanup to kill process
	p.RegisterCleanup(func(cleanupCtx context.Context) error {
		if cmd.Process != nil {
			if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
				return nil
			}
			if err := cmd.Process.Kill(); err != nil {
				if stderrors.Is(err, os.ErrProcessDone) {
					return nil
				}
				return err
			}
		}
		return nil
	})

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	ctx.Logger.Info("Starting transcription", map[string]interface{}{
		"model":      config.ModelSize,
		"audio_file": audioPath,
		"language":   config.Language,
	})

	if err := cmd.Run(); err != nil {
		ctx.Logger.Error("Whisper failed", map[string]interface{}{
			"error":  err.Error(),
			"stderr": stderr.String(),
		})
		return "", "", nil, fmt.Errorf("whisper failed: %w\nStderr: %s", err, stderr.String())
	}

	// Parse output based on format
	if config.OutputFormat == "json" {
		return p.parseJSONOutput(outputDir, audioPath)
	}

	return p.parseTXTOutput(outputDir, audioPath)
}

// parseJSONOutput parses Whisper's JSON output
func (p *WhisperTranscriber) parseJSONOutput(outputDir, audioPath string) (string, string, []interface{}, error) {
	// Whisper outputs to <basename>.json
	basename := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
	outputFile := filepath.Join(outputDir, basename+".json")

	data, err := os.ReadFile(outputFile)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to read whisper output: %w", err)
	}

	var whisperOut WhisperOutput
	if err := json.Unmarshal(data, &whisperOut); err != nil {
		return "", "", nil, fmt.Errorf("failed to parse whisper output: %w", err)
	}

	// Convert segments to interface{} for output
	segments := make([]interface{}, 0, len(whisperOut.Segments))
	for _, seg := range whisperOut.Segments {
		segments = append(segments, map[string]interface{}{
			"id":    seg.ID,
			"start": seg.Start,
			"end":   seg.End,
			"text":  seg.Text,
		})
	}

	return whisperOut.Text, whisperOut.Language, segments, nil
}

// parseTXTOutput parses Whisper's TXT output
func (p *WhisperTranscriber) parseTXTOutput(outputDir, audioPath string) (string, string, []interface{}, error) {
	// Whisper outputs to <basename>.txt
	basename := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
	outputFile := filepath.Join(outputDir, basename+".txt")

	data, err := os.ReadFile(outputFile)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to read whisper output: %w", err)
	}

	transcript := string(data)

	// TXT format doesn't include language or segments
	return transcript, "", nil, nil
}

// Helper functions

// checkWhisperInstalled checks if whisper is available in PATH
func checkWhisperInstalled() error {
	_, err := exec.LookPath("whisper")
	if err != nil {
		return fmt.Errorf("whisper not found in PATH")
	}
	return nil
}
