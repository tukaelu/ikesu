package logger

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type Logger struct {
	Log slog.Logger
}

// NewLogger returns a logger.
func NewLogger(file, level string, dryrun bool) (*Logger, error) {
	var logger *slog.Logger
	var logLevel slog.Level

	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, err
	}
	logOpt := slog.HandlerOptions{
		Level: &logLevel,
	}

	if file == "" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &logOpt))
	} else {
		fw, err := newFileWriter(file)
		if err != nil {
			return nil, err
		}
		logger = slog.New(slog.NewJSONHandler(fw, &logOpt))
	}

	return &Logger{Log: *logger}, nil
}

func newFileWriter(file string) (io.Writer, error) {
	if fi, err := os.Stat(file); err == nil && fi.IsDir() {
		return nil, errors.New("file is directory")
	}
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
}
