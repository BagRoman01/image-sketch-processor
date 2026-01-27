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
	Host string `yaml:"host" envconfig:"app_host"`
	Port int    `yaml:"port" envconfig:"app_port"`
}

type ServiceConfig struct {
	baseCfg.BaseServiceConfig
	InstanceConfig InstanceConfig `yaml:"instance"`
}

func NewServiceConfig() *ServiceConfig {
	cfg := &ServiceConfig{
		InstanceConfig: InstanceConfig{
			Host: "0.0.0.0",
			Port: 80,
		},
	}
	return cfg.Fill()
}

func (c *ServiceConfig) Fill() *ServiceConfig {
	if err := defaults.Set(&c.BaseServiceConfig); err != nil {
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

func loadYaml(c *ServiceConfig) error {
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
