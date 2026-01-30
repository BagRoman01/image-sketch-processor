package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type InstanceConfig struct {
	Host string `yaml:"host" envconfig:"host"`
	Port int    `yaml:"port" envconfig:"port"`
}

func NewInstanceConfig() *InstanceConfig {
	return &InstanceConfig{
		Host: "0.0.0.0",
		Port: 80,
	}
}

type RabbitMQConfig struct {
	URL       string `yaml:"url" envconfig:"rabbitmq_url"`
	QueueName string `yaml:"queue_name" envconfig:"rabbitmq_queue_name"`
}

func NewRabbitMQConfig() *RabbitMQConfig {
	return &RabbitMQConfig{
		URL:       "amqp://guest:guest@localhost:5672/",
		QueueName: "file_processing_tasks",
	}
}

type RedisConfig struct {
	Addr            string `yaml:"addr" envconfig:"redis_addr"`
	Password        string `yaml:"password" envconfig:"redis_password"`
	DB              int    `yaml:"db" envconfig:"redis_db"`
	MaxRetries      int    `yaml:"max_retries" envconfig:"redis_max_retries"`
	DialTimeoutSec  int    `yaml:"dial_timeout_sec" envconfig:"redis_dial_timeout"`
	ReadTimeoutSec  int    `yaml:"read_timeout_sec" envconfig:"redis_read_timeout"`
	WriteTimeoutSec int    `yaml:"write_timeout_sec" envconfig:"redis_write_timeout"`
	PoolSize        int    `yaml:"pool_size" envconfig:"redis_pool_size"`
}

func NewRedisConfig() *RedisConfig {
	return &RedisConfig{
		Addr:            "localhost:6379",
		Password:        "",
		DB:              0,
		MaxRetries:      3,
		DialTimeoutSec:  5,
		ReadTimeoutSec:  3,
		WriteTimeoutSec: 3,
		PoolSize:        10,
	}
}

type S3StorageConfig struct {
	Region            string `yaml:"region" envconfig:"s3_region"`
	Bucket            string `yaml:"bucket" envconfig:"s3_bucket"`
	AccessKeyID       string `yaml:"access_key_id" envconfig:"s3_access_key_id"`
	SecretAccessKey   string `yaml:"secret_access_key" envconfig:"s3_secret_access_key"`
	Endpoint          string `yaml:"endpoint" envconfig:"s3_endpoint"`
	UseSSL            bool   `yaml:"use_ssl" envconfig:"s3_use_ssl"`
	MaxUploadSize     int64  `yaml:"max_upload_size" envconfig:"s3_max_upload_size"`
	ChunkUploadSize   int64  `yaml:"chunk_upload_size" envconfig:"s3_chunk_upload_size"`
	UploadConcurrency uint16 `yaml:"upload_concurrency" envconfig:"upload_concurrency"`
}

func NewS3StorageConfig() *S3StorageConfig {
	return &S3StorageConfig{
		Region:            "ru-central1",
		Bucket:            "files",
		AccessKeyID:       "s3_access_key_id",
		SecretAccessKey:   "s3_secret_access_key",
		Endpoint:          "http://s3-storage:9000",
		UseSSL:            false,
		MaxUploadSize:     104857600, // 100 MB
		ChunkUploadSize:   5242880,   // 5 MB
		UploadConcurrency: 5,
	}
}

type Config struct {
	InstanceConfig  InstanceConfig  `yaml:"instance"`
	S3StorageConfig S3StorageConfig `yaml:"s3storage"`
	RedisConfig     RedisConfig     `yaml:"redis"`
	RabbitMQConfig  RabbitMQConfig  `yaml:"rabbitMQ"`
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

	slog.Info("loading configuration", "config_path", cfg.ConfigPath)
	return cfg.Fill()
}

func (c *Config) Fill() *Config {
	if err := loadYaml(c); err != nil {
		slog.Error("failed to load YAML config", "error", err, "path", c.ConfigPath)
		panic(err)
	}

	if err := envconfig.Process("", c); err != nil {
		slog.Error("failed to process environment variables", "error", err)
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
		slog.Warn("config file not found, using defaults", "path", c.ConfigPath)
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	file, err := os.Open(c.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	slog.Debug("parsing YAML config", "path", c.ConfigPath)

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(c); err != nil {
		return fmt.Errorf("failed to decode YAML config: %w", err)
	}

	return nil
}
