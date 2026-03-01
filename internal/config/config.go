package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type Config struct {
	InstanceConfig  InstanceConfig  `yaml:"instance"`
	S3StorageConfig S3StorageConfig `yaml:"s3storage"`
	RedisConfig     RedisConfig     `yaml:"redis"`
	RabbitMQConfig  RabbitMQConfig  `yaml:"rabbitmq"`
	LogConfig       LogConfig       `yaml:"logging"`
	ConfigPath      string          `envconfig:"config_path"`
}

func NewConfig() *Config {
	cfg := &Config{
		InstanceConfig:  *NewInstanceConfig(),
		S3StorageConfig: *NewS3StorageConfig(),
		RedisConfig:     *NewRedisConfig(),
		RabbitMQConfig:  *NewRabbitMQConfig(),
		LogConfig:       *NewLogConfig(),
		ConfigPath:      "config.yaml",
	}

	slog.Info(
		"loading configuration",
		"config_path", cfg.ConfigPath,
	)
	return cfg.Fill()
}

func (c *Config) Fill() *Config {
	if err := loadYaml(c); err != nil {
		slog.Error(
			"failed to load YAML config",
			"error", err,
			"path", c.ConfigPath,
		)
		panic(err)
	}

	if err := envconfig.Process("", c); err != nil {
		slog.Error(
			"failed to process environment variables",
			"error", err,
		)
		panic(err)
	}

	slog.Info("configuration loaded successfully",
		"host", c.InstanceConfig.Host,
		"port", c.InstanceConfig.Port,
		"s3_endpoint", c.S3StorageConfig.Endpoint,
		"s3_bucket", c.S3StorageConfig.Bucket,
	)
	return c
}

func loadYaml(c *Config) error {
	if _, err := os.Stat(c.ConfigPath); os.IsNotExist(err) {
		slog.Warn(
			"config file not found, using defaults",
			"path", c.ConfigPath,
		)
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	file, err := os.Open(c.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(c); err != nil {
		return fmt.Errorf("failed to decode YAML config: %w", err)
	}

	return nil
}
