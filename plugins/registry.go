package plugins

import (
	"github.com/yourusername/geee/internal/core"
	httpfetcher "github.com/yourusername/geee/plugins/http-fetcher"
	jsontransformer "github.com/yourusername/geee/plugins/json-transformer"
	regexextractor "github.com/yourusername/geee/plugins/regex-extractor"
	templaterenderer "github.com/yourusername/geee/plugins/template-renderer"
)

// RegisterAll registers all available plugins with the provided registry
func RegisterAll(registry core.PluginRegistry) error {
	plugins := []core.Plugin{
		jsontransformer.New(),
		regexextractor.New(),
		httpfetcher.New(),
		templaterenderer.New(),
	}

	for _, plugin := range plugins {
		if err := registry.Register(plugin); err != nil {
			return err
		}
	}

	return nil
}
