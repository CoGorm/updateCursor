package main

import (
	"os"

	"github.com/CoGorm/updateCursor/internal/cli"
)

func main() {
	// Get command line arguments (skip the program name)
	args := os.Args[1:]

	// Run the CLI application
	if err := cli.Run(args); err != nil {
		// Check if this is an "update needed" error
		if err.Error() == "update needed" {
			os.Exit(10) // Exit with status 10 for update needed (matching bash script behavior)
		}
		// Other errors exit with status 1
		os.Exit(1)
	}
}
