package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/geee/internal/cli"
)

var (
	// Version is set during build
	Version = "dev"
	// BuildTime is set during build
	BuildTime = "unknown"
)

const (
	usageText = `GEEE - Generic Extraction & Enrichment Engine

Usage:
  geee [command] [flags]

Commands:
  run             Execute a pipeline with the given configuration
  plugins list    List all registered plugins
  help            Show this help message

Flags for 'run' command:
  --config string   Path to the pipeline configuration file (required)
  --input string    Path to the input data file (JSON or YAML)
  --output string   Path to save the output result (JSON or YAML)
  --verbose         Enable verbose logging with debug information

Examples:
  # Run a pipeline with configuration and input/output files
  geee run --config configs/pipeline.yaml --input data.json --output result.json

  # Run with verbose logging
  geee run --config configs/pipeline.yaml --input data.json --verbose

  # Run without input (empty state)
  geee run --config configs/pipeline.yaml --output result.json

  # Output to stdout (no --output flag)
  geee run --config configs/pipeline.yaml --input data.json

  # List all registered plugins
  geee plugins list

Version: %s
Build Time: %s
`
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "run":
		runCommand()
	case "plugins":
		pluginsCommand()
	case "help", "--help", "-h":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// runCommand handles the 'run' command
func runCommand() {
	runFlags := flag.NewFlagSet("run", flag.ExitOnError)
	configPath := runFlags.String("config", "", "Path to the pipeline configuration file (required)")
	inputPath := runFlags.String("input", "", "Path to the input data file (JSON or YAML)")
	outputPath := runFlags.String("output", "", "Path to save the output result (JSON or YAML)")
	verbose := runFlags.Bool("verbose", false, "Enable verbose logging with debug information")

	runFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: geee run [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		runFlags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  geee run --config configs/pipeline.yaml --input data.json --output result.json\n")
		fmt.Fprintf(os.Stderr, "  geee run --config configs/pipeline.yaml --input data.json --verbose\n")
	}

	if err := runFlags.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Validate required flags
	if *configPath == "" {
		fmt.Fprintf(os.Stderr, "Error: --config flag is required\n\n")
		runFlags.Usage()
		os.Exit(1)
	}

	// Create CLI instance
	c := cli.NewCLI(*verbose)

	// Run the pipeline
	if err := c.Run(*configPath, *inputPath, *outputPath, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// pluginsCommand handles the 'plugins' command
func pluginsCommand() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: geee plugins <subcommand>\n\n")
		fmt.Fprintf(os.Stderr, "Subcommands:\n")
		fmt.Fprintf(os.Stderr, "  list    List all registered plugins\n")
		os.Exit(1)
	}

	subcommand := os.Args[2]

	switch subcommand {
	case "list":
		c := cli.NewCLI(false)
		if err := c.ListPlugins(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

// printUsage prints the usage information
func printUsage() {
	fmt.Printf(usageText, Version, BuildTime)
}
