package config

type BaseConfig struct {
	ConfigPath string    `envconfig:"config_path" required:"false" default:"config.yaml"`
	Logger     LogConfig `yaml:"logging"`
}
