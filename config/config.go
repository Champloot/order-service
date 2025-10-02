package config

import (
	"os"
	"strconv"
	"time"
	"strings"
)

type Config struct {
	PostgresConnStr string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	CacheTTL        time.Duration
	KafkaBrokers    []string
	KafkaTopic      string
	KafkaGroupID    string
	HTTPAddr        string
}

func LoadConfig() *Config {
	return &Config{
		PostgresConnStr: getEnv("POSTGRES_CONN_STR", "postgres://user:password@localhost:5432/orderservice?sslmode=disable"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:         getEnvAsInt("REDIS_DB", 0),
		CacheTTL:        getEnvAsDuration("CACHE_TTL", 24*time.Hour),
		KafkaBrokers:    getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}, ","),
		KafkaTopic:      getEnv("KAFKA_TOPIC", "orders"),
		KafkaGroupID:    getEnv("KAFKA_GROUP_ID", "order-service"),
		HTTPAddr:        getEnv("HTTP_ADDR", ":8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string, separator string) []string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.Split(value, separator)
	}
	return defaultValue
}
