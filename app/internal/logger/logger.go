package logger

import (
	"log/slog"
	"manimatic/internal/config"
	"os"
)

func NewLogger(cfg *config.Config) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: cfg.Logging.Level,
	}

	switch cfg.Logging.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
