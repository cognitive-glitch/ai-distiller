package main

import (
	"os"

	"github.com/janreges/ai-distiller/internal/cli"
)

// Version is set at build time via ldflags
var version = "dev"

func main() {
	// Set version for CLI
	cli.Version = version

	// Execute the root command
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}