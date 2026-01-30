package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BagRoman01/image-sketch-processor/internal/config"
)

func InitLogger(cfg *config.LogConfig) (*slog.Logger, error) {
	var output io.Writer
	switch cfg.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	case "file":
		if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		output = file
	default:
		return nil, fmt.Errorf("invalid log output: %s", cfg.Output)
	}

	level, err := cfg.ParseLevel()
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				a.Key = "ts"
				if t, ok := a.Value.Any().(time.Time); ok {
					switch cfg.TimeFormat {
					case "unix":
						a.Value = slog.Int64Value(t.Unix())
					case "iso8601":
						a.Value = slog.StringValue(t.UTC().Format("2006-01-02T15:04:05.000Z07:00"))
					default: // rfc3339
						a.Value = slog.StringValue(t.UTC().Format(time.RFC3339))
					}
				}
			case slog.LevelKey:
				a.Key = "lvl"
			case slog.MessageKey:
				a.Key = "msg"
			case slog.SourceKey:
				a.Key = "src"
				if source, ok := a.Value.Any().(*slog.Source); ok {
					funcName := source.Function
					if idx := strings.LastIndex(funcName, "/"); idx != -1 {
						funcName = funcName[idx+1:]
					}
					a.Value = slog.StringValue(fmt.Sprintf("%s:%d", funcName, source.Line))
				}
			}

			switch a.Key {
			case "request_id":
				a.Key = "req_id"
			case "duration_ms":
				a.Key = "dur_ms"
			}

			return a
		},
	}

	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	case "text":
		handler = slog.NewTextHandler(output, opts)
	default:
		return nil, fmt.Errorf("invalid log format: %s", cfg.Format)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger, nil
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, "request_id", requestID)
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	logger := slog.Default()

	if requestID, ok := ctx.Value("request_id").(string); ok {
		logger = logger.With("request_id", requestID)
	}

	return logger
}
