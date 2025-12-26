// Package cmd implements the command-line interface for fdups using Cobra.
//
// The CLI provides subcommands for various duplicate file detection operations.
// Currently supported commands:
//   - scan: Scan a directory for duplicate files
//
// Usage:
//
//	fdups <command> [flags] [arguments]
//
// Use "fdups <command> --help" for more information about a command.
package cmd

import (
	"os"

	"fdups/log"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// rootCmd is the base command for the fdups CLI.
var rootCmd = &cobra.Command{
	Use:   "fdups",
	Short: "Find duplicate files by content hash",
	Long:  "fdups is a CLI tool that finds duplicate files by computing and comparing content hashes.",
}

// Execute runs the root command and handles any errors.
// This is the main entry point for the CLI, called from main().
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.L().Fatal("Command execution failed", zap.Error(err))
		os.Exit(1)
	}
}
