package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	Log slog.Logger
}

// NewDefaultLogger returns a logger.
func NewDefaultLogger(level string, dryrun bool) (*Logger, error) {
	var logger *slog.Logger
	var logLevel slog.Level

	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, err
	}
	logOpt := slog.HandlerOptions{
		Level: &logLevel,
	}
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &logOpt))

	return &Logger{Log: *logger}, nil
}
