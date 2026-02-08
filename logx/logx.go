package logx

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xiaowumin-mark/AMLX/config"
)

var (
	mu            sync.RWMutex
	defaultLogger *slog.Logger
	defaultWriter io.Writer = os.Stdout
	closer        io.Closer
)

func Init(cfg config.LogConfig) (*slog.Logger, error) {
	writer, c, err := buildWriter(cfg)
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{
		Level:     parseLevel(cfg.Level),
		AddSource: cfg.AddSource,
	}
	if cfg.TimeFormat != "" {
		opts.ReplaceAttr = func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				if t, ok := attr.Value.Any().(time.Time); ok {
					attr.Value = slog.StringValue(t.Format(cfg.TimeFormat))
				}
			}
			return attr
		}
	}

	var handler slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	default:
		handler = slog.NewTextHandler(writer, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	mu.Lock()
	if closer != nil {
		_ = closer.Close()
	}
	closer = c
	defaultLogger = logger
	defaultWriter = writer
	mu.Unlock()

	return logger, nil
}

func L() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	if defaultLogger != nil {
		return defaultLogger
	}
	return slog.Default()
}

func Writer() io.Writer {
	mu.RLock()
	defer mu.RUnlock()
	if defaultWriter != nil {
		return defaultWriter
	}
	return os.Stdout
}

func Close() error {
	mu.Lock()
	defer mu.Unlock()
	if closer == nil {
		return nil
	}
	err := closer.Close()
	closer = nil
	return err
}

func buildWriter(cfg config.LogConfig) (io.Writer, io.Closer, error) {
	output := strings.ToLower(strings.TrimSpace(cfg.Output))
	switch output {
	case "", "stdout":
		return os.Stdout, nil, nil
	case "stderr":
		return os.Stderr, nil, nil
	case "discard", "none", "off":
		return io.Discard, nil, nil
	case "file", "both", "stdout+file", "file+stdout":
		if cfg.File == "" {
			return nil, nil, errors.New("log.file is required when log.output is file/both")
		}
		path, err := filepath.Abs(cfg.File)
		if err != nil {
			return nil, nil, err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, nil, err
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, nil, err
		}
		if output == "file" {
			return f, f, nil
		}
		return io.MultiWriter(os.Stdout, f), f, nil
	default:
		return nil, nil, errors.New("log.output must be one of stdout|stderr|file|both|discard")
	}
}

func parseLevel(level string) slog.Leveler {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info", "":
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}

func Fatal(msg string, err error) {
	L().Error(msg, "error", err)
	os.Exit(1)
}
