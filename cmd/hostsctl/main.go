// hostsctl is a command-line tool for safely managing entries in /etc/hosts files.
//
// It provides functionality to:
//   - List, add, remove, enable/disable hosts entries
//   - Create and restore backups with timestamps
//   - Import/export profiles in JSON/YAML format
//   - Validate hosts file syntax and detect issues
//   - Perform atomic writes with file locking for safety
//
// Usage examples:
//
//	hostsctl list                                    # Show active entries
//	hostsctl add --ip 127.0.0.1 --name app.local    # Add new entry
//	hostsctl backup                                  # Create backup
//	hostsctl verify                                  # Check file validity
//
// For complete usage information, run: hostsctl --help
package main

import (
	"fmt"
	"os"

	"github.com/vaxvhbe/hostsctl/internal/cli"
)

// main is the entry point for the hostsctl application.
// It creates a CLI instance and executes the user's command.
func main() {
	app := cli.NewCLI()

	if err := app.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
