package config

import (
	"fmt"
	"os"

	baseCfg "github.com/BagRoman01/image-sketch-processor/pkg/base/config"
	"github.com/creasty/defaults"
	"github.com/kelseyhightower/envconfig"
	"go.yaml.in/yaml/v3"
)

type InstanceConfig struct {
	Host string `yaml:"host" envconfig:"host"`
	Port int    `yaml:"port" envconfig:"port"`
}

type Config struct {
	baseCfg.BaseConfig
	InstanceConfig  InstanceConfig  `yaml:"instance"`
	S3StorageConfig S3StorageConfig `yaml:"s3storage"`
}

type S3StorageConfig struct {
	Region            string `yaml:"region" envconfig:"s3_region" default:"ru-central1"`
	Bucket            string `yaml:"bucket" envconfig:"s3_bucket" default:"files"`
	AccessKeyID       string `yaml:"access_key_id" envconfig:"s3_access_key_id"`
	SecretAccessKey   string `yaml:"secret_access_key" envconfig:"s3_secret_access_key"`
	Endpoint          string `yaml:"endpoint" envconfig:"s3_endpoint"`
	UseSSL            bool   `yaml:"use_ssl" envconfig:"s3_use_ssl" default:"false"`
	MaxUploadSize     int64  `yaml:"max_upload_size" envconfig:"s3_max_upload_size"`
	ChunkUploadSize   int64  `yaml:"chunk_upload_size" envconfig:"s3_chunk_upload_size"`
	UploadConcurrency uint16 `yaml:"upload_concurrency" envconfig:"upload_concurrency"`
}

func NewConfig() *Config {
	cfg := &Config{
		InstanceConfig: InstanceConfig{
			Host: "0.0.0.0",
			Port: 80,
		},
		S3StorageConfig: S3StorageConfig{
			AccessKeyID:       "s3_access_key_id",
			SecretAccessKey:   "s3_secret_access_key",
			Endpoint:          "http://s3-storage:9000",
			MaxUploadSize:     104857600,
			ChunkUploadSize:   5242880,
			UploadConcurrency: 5,
		},
	}
	return cfg.Fill()
}

func (c *Config) Fill() *Config {
	if err := defaults.Set(&c.BaseConfig); err != nil {
		panic(err)
	}

	if err := loadYaml(c); err != nil {
		panic(err)
	}

	if err := envconfig.Process("", c); err != nil {
		panic(err)
	}
	return c
}

func loadYaml(c *Config) error {
	if _, err := os.Stat(c.ConfigPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("ошибка проверки существования файла: %w", err)
	}

	file, err := os.Open(c.ConfigPath)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}

	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(&c); err != nil {
		return fmt.Errorf("ошибка декодирования YAML конфига: %w", err)
	}

	return nil
}
