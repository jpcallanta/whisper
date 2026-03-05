package main

import (
	"os"
)

// Runs the root CLI and exits with non-zero on error.
func main() {
	// Execute failed.
	if err := Execute(); err != nil {
		os.Exit(1)
	}
}
