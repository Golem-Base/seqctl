package log

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/MatusOllah/slogcolor"
)

// InitForTUI initializes logging for TUI mode with file output
func InitForTUI(levelStr string, format string, noColor bool, logFile string) error {
	var output io.Writer

	if logFile == "" {
		// For TUI mode, if no log file specified, disable most logging
		// by setting a very high level and using discard
		output = io.Discard
		levelStr = "error" // Only errors will be logged to discard
	} else {
		// Open log file for writing
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			return err
		}
		output = file
	}

	Init(levelStr, format, noColor, output)
	return nil
}

// Init initializes the global logger with the specified level and output format
func Init(levelStr string, format string, noColor bool, output io.Writer) {
	// Set log level
	var level slog.Level
	switch strings.ToLower(levelStr) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Configure handler based on format
	var handler slog.Handler
	switch strings.ToLower(format) {
	case "json":
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{Level: level})
	default:
		if !noColor {
			opts := &slogcolor.Options{
				Level:   level,
				NoColor: false,
			}
			handler = slogcolor.NewHandler(output, opts)
		} else {
			handler = slog.NewTextHandler(output, &slog.HandlerOptions{Level: level})
		}
	}

	// Set the default logger
	slog.SetDefault(slog.New(handler))
}
