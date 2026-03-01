package config

import "fmt"

type InstanceConfig struct {
	Host string `yaml:"host" envconfig:"host"`
	Port int    `yaml:"port" envconfig:"port"`
}

func (instance *InstanceConfig) Address() string {
	return fmt.Sprintf(
		"%s:%d",
		instance.Host,
		instance.Port,
	)
}

func NewInstanceConfig() *InstanceConfig {
	return &InstanceConfig{
		Host: "0.0.0.0",
		Port: 80,
	}
}
