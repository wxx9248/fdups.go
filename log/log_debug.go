//go:build debug

// Package log provides a singleton logger for the application.
//
// The logger implementation varies based on build tags:
//   - debug: Uses zap's development configuration with verbose output
//   - release: Uses a custom production configuration with minimal output
//
// Access the logger via the L() function.
package log

import (
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	logger = zap.Must(zap.NewDevelopment())
}

// L returns the singleton zap logger instance.
//
// In debug builds, this returns a development logger with:
//   - Human-readable console output
//   - Debug level enabled
//   - Caller information included
func L() *zap.Logger {
	return logger
}
