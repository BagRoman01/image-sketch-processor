package config

type RabbitMQConfig struct {
	URL       string `yaml:"url"   envconfig:"rabbitmq_url"`
	QueueName string `yaml:"queue" envconfig:"rabbitmq_queue_name"`
}

func NewRabbitMQConfig() *RabbitMQConfig {
	return &RabbitMQConfig{
		URL:       "amqp://guest:guest@localhost:5672/",
		QueueName: "file_processing_tasks",
	}
}
