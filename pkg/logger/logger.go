// Package logger sets up structured logging using slog with dev or production modes.
// In dev mode, logs are printed to stdout. In production, logs are written to a file.

package logger

import (
	"log/slog"
	"os"
)

type Mode string

const (
	ModeDev  Mode = "dev"  // Development mode (stdout logging)
	ModeProd Mode = "prod" // Production mode (file logging)
)

// LoggerOptions defines configuration for the logger.
type LoggerOptions struct {
	Mode Mode   // Logging mode (dev or prod)
	File string // File path for production logs
}

// NewLogger initializes the global logger based on the provided options.
// In dev mode, logs go to stdout. In prod mode, logs are written to the specified file.
func NewLogger(loggerOpts LoggerOptions) error {
	opts := PrettyHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: slog.LevelInfo,
		},
	}

	var out = os.Stdout
	if loggerOpts.Mode == ModeProd {
		file, err := os.OpenFile(loggerOpts.File, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		out = file
	}

	handler := NewPrettyHandler(out, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return nil
}
