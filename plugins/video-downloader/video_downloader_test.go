package videodownloader

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
	assert.Equal(t, "video-downloader", plugin.Manifest().ID)
	assert.Equal(t, "1.0.0", plugin.Manifest().Version)
	assert.Equal(t, 300, plugin.Manifest().Timeout)
}

func TestVideoDownloader_ManifestSchemas(t *testing.T) {
	plugin := New()
	manifest := plugin.Manifest()

	// Verify input schema
	assert.NotNil(t, manifest.InputSchema)
	assert.Contains(t, manifest.InputSchema, "properties")

	// Verify output schema
	assert.NotNil(t, manifest.OutputSchema)
	outputProps := manifest.OutputSchema["properties"].(map[string]interface{})
	assert.Contains(t, outputProps, "video_file")
	assert.Contains(t, outputProps, "audio_file")
	assert.Contains(t, outputProps, "duration")
	assert.Contains(t, outputProps, "metadata")

	// Verify config schema
	assert.NotNil(t, manifest.ConfigSchema)
	configProps := manifest.ConfigSchema["properties"].(map[string]interface{})
	assert.Contains(t, configProps, "video_url")
	assert.Contains(t, configProps, "output_format")
	assert.Contains(t, configProps, "quality")
	assert.Contains(t, configProps, "extract_audio")
}

func TestVideoDownloader_Execute_NoURL(t *testing.T) {
	plugin := New()

	config := map[string]interface{}{}
	input := map[string]interface{}{} // No video_url

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
	assert.Contains(t, err.Error(), "video_url is required")
}

func TestVideoDownloader_Execute_URLFromConfig(t *testing.T) {
	// Skip if yt-dlp not installed
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		t.Skip("yt-dlp not installed, skipping integration test")
	}

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	plugin := New()

	// Use a very short test video
	testURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ" // Rick Astley - Never Gonna Give You Up (short clip)

	config := map[string]interface{}{
		"video_url":     testURL,
		"output_format": "mp4",
		"quality":       "worst", // Use worst quality for faster download
		"extract_audio": true,
	}

	input := map[string]interface{}{}

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
	assert.Contains(t, result.Data, "video_file")
	assert.Contains(t, result.Data, "audio_file")
	assert.Contains(t, result.Data, "duration")
	assert.Contains(t, result.Data, "metadata")

	// Verify file:// references
	videoRef := result.Data["video_file"].(string)
	assert.True(t, len(videoRef) > 7 && videoRef[:7] == "file://")

	audioRef := result.Data["audio_file"].(string)
	assert.True(t, len(audioRef) > 7 && audioRef[:7] == "file://")

	// Cleanup
	err = plugin.Cleanup(context.Background())
	assert.NoError(t, err)
}

func TestVideoDownloader_Execute_URLFromState(t *testing.T) {
	// Skip if yt-dlp not installed
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		t.Skip("yt-dlp not installed, skipping integration test")
	}

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	plugin := New()

	testURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

	config := map[string]interface{}{
		"quality":       "worst",
		"extract_audio": false, // Don't extract audio to speed up test
	}

	input := map[string]interface{}{
		"video_url": testURL, // URL in state
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

	// State should be preserved
	assert.Equal(t, testURL, result.Data["video_url"])

	// Cleanup
	err = plugin.Cleanup(context.Background())
	assert.NoError(t, err)
}

func TestCheckYtDlpInstalled(t *testing.T) {
	err := checkYtDlpInstalled()
	if err != nil {
		t.Log("yt-dlp not installed (this is OK for unit tests)")
	}
}

func TestFindVideoFile(t *testing.T) {
	// Create temp directory with test files
	tempDir := t.TempDir()

	// Create video file
	videoPath := filepath.Join(tempDir, "video.mp4")
	err := os.WriteFile(videoPath, []byte("fake video"), 0644)
	require.NoError(t, err)

	// Create info.json (should be ignored)
	jsonPath := filepath.Join(tempDir, "video.info.json")
	err = os.WriteFile(jsonPath, []byte("{}"), 0644)
	require.NoError(t, err)

	// Test finding video file
	found, err := findVideoFile(tempDir)
	assert.NoError(t, err)
	assert.Equal(t, videoPath, found)
}

func TestFindVideoFile_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	_, err := findVideoFile(tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no video file found")
}

func TestParseMetadata(t *testing.T) {
	tempDir := t.TempDir()

	// Create mock info.json
	infoPath := filepath.Join(tempDir, "video.info.json")
	infoJSON := `{
		"title": "Test Video",
		"uploader": "Test User",
		"upload_date": "20260113",
		"extractor": "tiktok",
		"duration": 45.5
	}`
	err := os.WriteFile(infoPath, []byte(infoJSON), 0644)
	require.NoError(t, err)

	metadata, duration := parseMetadata(tempDir)

	assert.Equal(t, "Test Video", metadata["title"])
	assert.Equal(t, "Test User", metadata["uploader"])
	assert.Equal(t, "20260113", metadata["upload_date"])
	assert.Equal(t, "tiktok", metadata["platform"])
	assert.Equal(t, 45.5, duration)
}

func TestParseMetadata_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()

	metadata, duration := parseMetadata(tempDir)

	// Should return empty metadata
	assert.Empty(t, metadata)
	assert.Equal(t, 0.0, duration)
}

func TestGetString(t *testing.T) {
	m := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": nil,
	}

	assert.Equal(t, "value1", getString(m, "key1"))
	assert.Equal(t, "", getString(m, "key2")) // Not a string
	assert.Equal(t, "", getString(m, "key3")) // Nil
	assert.Equal(t, "", getString(m, "key4")) // Not exists
}
