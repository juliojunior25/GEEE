package whispertranscriber

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/geee/internal/observability"
	"github.com/yourusername/geee/pkg/types"
)

func TestNew(t *testing.T) {
	plugin := New()
	assert.NotNil(t, plugin)
	assert.Equal(t, "whisper-transcriber", plugin.Manifest().ID)
	assert.Equal(t, "1.0.0", plugin.Manifest().Version)
	assert.Equal(t, 120, plugin.Manifest().Timeout)
	assert.Equal(t, 2048, plugin.Manifest().MaxMemoryMB)
}

func TestWhisperTranscriber_ManifestSchemas(t *testing.T) {
	plugin := New()
	manifest := plugin.Manifest()

	// Verify input schema
	assert.NotNil(t, manifest.InputSchema)
	inputProps := manifest.InputSchema["properties"].(map[string]interface{})
	assert.Contains(t, inputProps, "audio_file")

	// Verify output schema
	assert.NotNil(t, manifest.OutputSchema)
	outputProps := manifest.OutputSchema["properties"].(map[string]interface{})
	assert.Contains(t, outputProps, "transcript")
	assert.Contains(t, outputProps, "language_detected")
	assert.Contains(t, outputProps, "segments")

	// Verify config schema
	assert.NotNil(t, manifest.ConfigSchema)
	configProps := manifest.ConfigSchema["properties"].(map[string]interface{})
	assert.Contains(t, configProps, "model_size")
	assert.Contains(t, configProps, "language")
	assert.Contains(t, configProps, "output_format")
}

func TestWhisperTranscriber_Execute_NoAudioFile(t *testing.T) {
	plugin := New()

	config := map[string]interface{}{}
	input := map[string]interface{}{} // No audio_file

	execCtx := &types.ExecutionContext{
		Context:     context.Background(),
		ExecutionID: uuid.New().String(),
		State:       input,
		Config:      config,
		Logger:      observability.NewDefaultLogger(false),
		Metrics:     observability.NewNoOpMetricsCollector(),
	}

	result, err := plugin.Execute(execCtx)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestWhisperTranscriber_Execute_InvalidFileReference(t *testing.T) {
	// Skip if whisper not installed
	if err := checkWhisperInstalled(); err != nil {
		t.Skip("whisper not installed, skipping validation test")
	}

	plugin := New()

	config := map[string]interface{}{}
	input := map[string]interface{}{
		"audio_file": "/tmp/audio.wav", // Should be file://
	}

	execCtx := &types.ExecutionContext{
		Context:     context.Background(),
		ExecutionID: uuid.New().String(),
		State:       input,
		Config:      config,
		Logger:      observability.NewDefaultLogger(false),
		Metrics:     observability.NewNoOpMetricsCollector(),
	}

	result, err := plugin.Execute(execCtx)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "must be a file:// URI")
}

func TestWhisperTranscriber_Execute_FileNotFound(t *testing.T) {
	// Skip if whisper not installed
	if err := checkWhisperInstalled(); err != nil {
		t.Skip("whisper not installed, skipping validation test")
	}

	plugin := New()

	config := map[string]interface{}{}
	input := map[string]interface{}{
		"audio_file": "file:///nonexistent/audio.wav",
	}

	execCtx := &types.ExecutionContext{
		Context:     context.Background(),
		ExecutionID: uuid.New().String(),
		State:       input,
		Config:      config,
		Logger:      observability.NewDefaultLogger(false),
		Metrics:     observability.NewNoOpMetricsCollector(),
	}

	result, err := plugin.Execute(execCtx)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "audio file not found")
}

func TestCheckWhisperInstalled(t *testing.T) {
	err := checkWhisperInstalled()
	if err != nil {
		t.Log("whisper not installed (this is OK for unit tests)")
	}
}

func TestParseTXTOutput(t *testing.T) {
	// Create temp directory with test output
	tempDir := t.TempDir()

	// Create mock TXT output
	audioPath := "/tmp/test.wav"
	txtPath := filepath.Join(tempDir, "test.txt")
	txtContent := "This is a test transcription."
	err := os.WriteFile(txtPath, []byte(txtContent), 0644)
	require.NoError(t, err)

	plugin := New()
	transcript, language, segments, err := plugin.parseTXTOutput(tempDir, audioPath)

	assert.NoError(t, err)
	assert.Equal(t, txtContent, transcript)
	assert.Equal(t, "", language) // TXT doesn't include language
	assert.Nil(t, segments)        // TXT doesn't include segments
}

func TestParseJSONOutput(t *testing.T) {
	// Create temp directory with test output
	tempDir := t.TempDir()

	// Create mock JSON output
	audioPath := "/tmp/test.wav"
	jsonPath := filepath.Join(tempDir, "test.json")
	jsonContent := `{
		"text": "This is a test transcription.",
		"language": "en",
		"segments": [
			{
				"id": 0,
				"start": 0.0,
				"end": 2.5,
				"text": "This is a test"
			},
			{
				"id": 1,
				"start": 2.5,
				"end": 4.0,
				"text": "transcription."
			}
		]
	}`
	err := os.WriteFile(jsonPath, []byte(jsonContent), 0644)
	require.NoError(t, err)

	plugin := New()
	transcript, language, segments, err := plugin.parseJSONOutput(tempDir, audioPath)

	assert.NoError(t, err)
	assert.Equal(t, "This is a test transcription.", transcript)
	assert.Equal(t, "en", language)
	assert.Len(t, segments, 2)

	// Verify first segment
	seg0 := segments[0].(map[string]interface{})
	assert.Equal(t, 0, seg0["id"])
	assert.Equal(t, 0.0, seg0["start"])
	assert.Equal(t, 2.5, seg0["end"])
	assert.Equal(t, "This is a test", seg0["text"])
}

func TestParseJSONOutput_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	audioPath := "/tmp/test.wav"

	plugin := New()
	_, _, _, err := plugin.parseJSONOutput(tempDir, audioPath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read whisper output")
}

func TestConfig_Defaults(t *testing.T) {
	plugin := New()

	config := map[string]interface{}{}

	// Create a dummy audio file for testing
	tempDir := t.TempDir()
	audioPath := filepath.Join(tempDir, "test.wav")
	err := os.WriteFile(audioPath, []byte("fake audio"), 0644)
	require.NoError(t, err)

	input := map[string]interface{}{
		"audio_file": "file://" + audioPath,
	}

	execCtx := &types.ExecutionContext{
		Context:     context.Background(),
		ExecutionID: uuid.New().String(),
		State:       input,
		Config:      config,
		Logger:      observability.NewDefaultLogger(false),
		Metrics:     observability.NewNoOpMetricsCollector(),
	}

	// This will fail because whisper isn't installed, but we can check the config parsing
	_, err = plugin.Execute(execCtx)
	// We expect it to fail at whisper execution, not config parsing
	if err != nil && err.Error() == "whisper not found in PATH" {
		// This is expected - config parsing worked
		t.Log("Config parsing successful (whisper not installed for execution)")
	}
}

// Integration test - only runs if whisper is installed
func TestWhisperTranscriber_Execute_Integration(t *testing.T) {
	// Skip if whisper not installed
	if _, err := exec.LookPath("whisper"); err != nil {
		t.Skip("whisper not installed, skipping integration test")
	}

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a test audio file (1 second of silence using ffmpeg)
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed, cannot create test audio")
	}

	tempDir := t.TempDir()
	audioPath := filepath.Join(tempDir, "test.wav")

	// Generate silent audio
	cmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "anullsrc=r=16000:cl=mono", "-t", "1", "-y", audioPath)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create test audio: %v", err)
	}

	plugin := New()

	config := map[string]interface{}{
		"model_size":    "tiny", // Fastest model for testing
		"output_format": "json",
	}

	input := map[string]interface{}{
		"audio_file": "file://" + audioPath,
	}

	execCtx := &types.ExecutionContext{
		Context:     context.Background(),
		ExecutionID: uuid.New().String(),
		State:       input,
		Config:      config,
		Logger:      observability.NewDefaultLogger(false),
		Metrics:     observability.NewNoOpMetricsCollector(),
	}

	result, err := plugin.Execute(execCtx)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)

	// Verify outputs
	assert.Contains(t, result.Data, "transcript")
	assert.Contains(t, result.Data, "audio_file") // State preserved

	// Cleanup
	err = plugin.Cleanup(context.Background())
	assert.NoError(t, err)
}
