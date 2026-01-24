package plugins

import (
	"github.com/yourusername/geee/internal/core"
	firecrawlsearch "github.com/yourusername/geee/plugins/firecrawl-search"
	geocodingenricher "github.com/yourusername/geee/plugins/geocoding-enricher"
	httpfetcher "github.com/yourusername/geee/plugins/http-fetcher"
	jsontransformer "github.com/yourusername/geee/plugins/json-transformer"
	llmanalyzer "github.com/yourusername/geee/plugins/llm-analyzer"
	regexextractor "github.com/yourusername/geee/plugins/regex-extractor"
	templaterenderer "github.com/yourusername/geee/plugins/template-renderer"
	videodownloader "github.com/yourusername/geee/plugins/video-downloader"
	whispertranscriber "github.com/yourusername/geee/plugins/whisper-transcriber"
)

// RegisterAll registers all available plugins with the provided registry
func RegisterAll(registry core.PluginRegistry) error {
	plugins := []core.Plugin{
		firecrawlsearch.New(),
		geocodingenricher.New(),
		httpfetcher.New(),
		jsontransformer.New(),
		llmanalyzer.New(),
		regexextractor.New(),
		templaterenderer.New(),
		videodownloader.New(),
		whispertranscriber.New(),
	}

	for _, plugin := range plugins {
		if err := registry.Register(plugin); err != nil {
			return err
		}
	}

	return nil
}
