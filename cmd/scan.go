package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"fdups/finder"
	"fdups/log"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// finderType holds the --finder flag value.
var finderType string

// scanCmd represents the scan command.
var scanCmd = &cobra.Command{
	Use:   "scan <directory>",
	Short: "Scan a directory for duplicate files",
	Long:  "Scan a directory recursively and find duplicate files based on content hash.",
	Args:  cobra.ExactArgs(1),
	Run:   runScan,
}

func init() {
	scanCmd.Flags().StringVar(&finderType, "finder", "default", "Finder type: default, flac")
	rootCmd.AddCommand(scanCmd)
}

// runScan is the main entry point for the scan command.
func runScan(cmd *cobra.Command, args []string) {
	directory := resolveDirectory(args[0])
	f := createFinder(finderType, directory)
	result := executeFinder(f, directory)
	outputResult(result)
}

// resolveDirectory converts a relative path to an absolute path.
func resolveDirectory(directory string) string {
	if path.IsAbs(directory) {
		return directory
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.L().Fatal("Failed to get current working directory", zap.Error(err))
	}
	return path.Join(cwd, path.Clean(directory))
}

// createFinder returns a Finder based on the specified type.
func createFinder(finderType, directory string) finder.Finder {
	switch finderType {
	case "default":
		return finder.NewDefaultFinder(directory)
	case "flac":
		return finder.NewFlacFinder(directory)
	default:
		log.L().Fatal("Unknown finder type",
			zap.String("type", finderType),
			zap.Strings("valid", []string{"default", "flac"}))
		return nil
	}
}

// executeFinder runs the finder and returns the results.
func executeFinder(f finder.Finder, directory string) map[string][]finder.FileInfo {
	log.L().Info("Program started", zap.String("target", directory))
	start := time.Now()

	err, result := f.Find()
	if err != nil {
		log.L().Fatal("Program terminated with error", zap.Error(err))
	}

	log.L().Info("Program completed successfully", zap.Duration("duration", time.Since(start)))
	return result
}

// outputResult marshals the result to JSON and prints it to stdout.
func outputResult(result map[string][]finder.FileInfo) {
	jsonResult, err := json.Marshal(result)
	if err != nil {
		log.L().Fatal("Failed to marshal result", zap.Error(err))
	}
	fmt.Println(string(jsonResult))
}
