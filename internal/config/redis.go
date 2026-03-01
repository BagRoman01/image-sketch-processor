package config

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
