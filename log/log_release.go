//go:build release

// Package log provides a singleton logger for the application.
//
// The logger implementation varies based on build tags:
//   - debug: Uses zap's development configuration with verbose output
//   - release: Uses a custom production configuration with minimal output
//
// Access the logger via the L() function.
package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	logConfig := zap.NewProductionConfig()
	logConfig.Encoding = "console"
	logConfig.EncoderConfig.LevelKey = ""
	logConfig.EncoderConfig.CallerKey = ""
	logConfig.EncoderConfig.FunctionKey = ""
	logConfig.EncoderConfig.StacktraceKey = ""
	logConfig.EncoderConfig.EncodeTime = func(time time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(time.Format("15:04:05.000"))
	}
	logConfig.EncoderConfig.EncodeDuration = func(Duration time.Duration, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(Duration.String())
	}
	logger = zap.Must(logConfig.Build())
}

// L returns the singleton zap logger instance.
//
// In release builds, this returns a production logger with:
//   - Console encoding with minimal decoration
//   - Timestamp format: HH:MM:SS.mmm
//   - No caller, function, or stacktrace information
func L() *zap.Logger {
	return logger
}
