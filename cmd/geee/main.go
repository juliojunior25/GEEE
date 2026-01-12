package main

import (
	"fmt"
	"os"
)

var (
	// Version is set during build
	Version = "dev"
	// BuildTime is set during build
	BuildTime = "unknown"
)

func main() {
	fmt.Printf("GEEE - Generic Extraction & Enrichment Engine\n")
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Println("\nCLI implementation coming in Phase 3...")
	os.Exit(0)
}
