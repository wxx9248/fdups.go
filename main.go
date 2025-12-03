package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"fdups/finder"
	"fdups/log"

	"go.uber.org/zap"
)

func main() {
	if len(os.Args) < 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s <directory>\n", os.Args[0])
		os.Exit(1)
	}

	directory := os.Args[1]
	if !path.IsAbs(directory) {
		cwd, err := os.Getwd()
		handleError(err)
		directory = path.Join(cwd, path.Clean(directory))
	}

	defaultFinder := finder.NewDefaultFinder(directory)
	log.L().Info("Program started", zap.String("target", directory))
	start := time.Now()
	err, result := defaultFinder.Find()
	handleError(err)
	log.L().Info("Program completed successfully", zap.Duration("duration", time.Since(start)))

	jsonResult, err := json.Marshal(result)
	handleError(err)
	fmt.Println(string(jsonResult))
}

func handleError(err error) {
	if err != nil {
		log.L().Fatal("Program terminated with error", zap.Error(err))
	}
}
