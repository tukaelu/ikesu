package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/go-logr/zapr"
)

type Logger struct {
	Log logr.Logger
}

// NewDefaultLogger returns a logger.
func NewDefaultLogger(logLevel string, dryrun bool) (*Logger, error) {
	var logger logr.Logger

	if dryrun {
		logger = stdr.New(log.Default())
	} else {
		conf := zap.NewProductionConfig()
		conf.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
		conf.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		// Cease the sampling of log output.
		// see. https://github.com/uber-go/zap/blob/master/FAQ.md#why-are-some-of-my-logs-missing
		conf.Sampling = nil

		var lv zapcore.Level
		if err := lv.UnmarshalText([]byte(logLevel)); err != nil {
			conf.Level = zap.NewAtomicLevel()
		} else {
			conf.Level = zap.NewAtomicLevelAt(lv)
		}
		zapLog, err := conf.Build()
		if err != nil {
			return nil, err
		}
		logger = zapr.NewLogger(zapLog)
	}
	return &Logger{Log: logger}, nil
}
