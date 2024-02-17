//go:build debug

package log

import (
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	logger = zap.Must(zap.NewDevelopment())
}

func L() *zap.Logger {
	return logger
}
