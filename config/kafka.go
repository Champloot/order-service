package config

type KafkaConfig struct {
	Brokers  []string
	Topic    string
	MinBytes int
	MaxBytes int
}

func NewKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "orders",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	}
}
