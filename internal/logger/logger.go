package logger

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger     *zap.Logger = zap.NewNop()
	globalLoggerOnce sync.Once
)

func Init(debug bool) {
	globalLoggerOnce.Do(func() {
		var err error
		if debug {
			globalLogger, err = zap.NewDevelopment()
		} else {
			cfg := zap.NewProductionConfig()
			cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

			globalLogger, err = cfg.Build()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to initialize logger - falling back to NoOp logger. Reason: %v\n", err)
			globalLogger = zap.NewNop()
			return
		}
	})
}

func Get() *zap.Logger {
	return globalLogger
}

func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}
