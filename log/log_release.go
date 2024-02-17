//go:build release

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

func L() *zap.Logger {
	return logger
}
