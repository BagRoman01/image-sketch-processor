package config

import (
	"fmt"
	"log/slog"
	"strings"
)

type LogConfig struct {
	Level      string `yaml:"level" envconfig:"log_level"`
	Format     string `yaml:"format" envconfig:"log_format"`
	Output     string `yaml:"output" envconfig:"log_output"`
	FilePath   string `yaml:"file_path" envconfig:"log_file_path"`
	AddSource  bool   `yaml:"add_source" envconfig:"log_add_source"`
	TimeFormat string `yaml:"time_format" envconfig:"log_time_format"`
}

func NewLogConfig() *LogConfig {
	return &LogConfig{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		FilePath:   "logs/app.log",
		AddSource:  false,
		TimeFormat: "rfc3339",
	}
}

func (c *LogConfig) ParseLevel() (slog.Leveler, error) {
	switch strings.ToLower(c.Level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return nil, fmt.Errorf("invalid log level: %s", c.Level)
	}
}
