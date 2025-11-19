package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Database	DatabaseConfig
	Cache		CacheConfig
	Kafka		KafkaConfig
	HTTP		HTTPConfig
	Consumer	ConsumerConfig
	App			AppConfig
    Validation	ValidationConfig
}

type DatabaseConfig struct {
	URL					string
	MaxConns          	int32
	MinConns         	int32
	MaxConnLifetime   	time.Duration
	MaxConnIdleTime   	time.Duration
	HealthCheckPeriod 	time.Duration
}

type CacheConfig struct {
	Addr		string
	Password 	string
	DB       	int
	TTL      	time.Duration
}

type KafkaConfig struct {
	Brokers	[]string
	Topic   string
	GroupID	string
}

type HTTPConfig struct {
	Addr	string
}

type ConsumerConfig struct {
	Timeout		time.Duration
	MinBytes   	int
	MaxBytes   	int
	MaxWait    	time.Duration
	RetryDelay 	time.Duration
}

type AppConfig struct {
	Env			string
	LogLevel 	string
}

type ValidationConfig struct {
    Enabled				bool
    RequireOrderUID		bool
    RequireTrackNumber	bool
    RequireEntry		bool
    RequireCustomerID	bool
    RequireDelivery		bool
    RequirePayment		bool
    RequireItems		bool
    
    ValidateEmail	bool
    ValidatePhone	bool
    ValidateAmounts	bool
    ValidateDates	bool
    
    MaxItems		int
    MinAmount		int
    MaxAmount		int
    MaxDateFuture	time.Duration
    MaxDatePast		time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		Database: DatabaseConfig{
			URL:               getEnv("POSTGRES_URL", "postgres://user:password@localhost:5432/orderservice?sslmode=disable"),
			MaxConns:          getEnvAsInt32("DB_MAX_CONNS", 20),
			MinConns:          getEnvAsInt32("DB_MIN_CONNS", 5),
			MaxConnLifetime:   getEnvAsDuration("DB_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime:   getEnvAsDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
			HealthCheckPeriod: getEnvAsDuration("DB_HEALTH_CHECK_PERIOD", time.Minute),
		},
		Cache: CacheConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			TTL:      getEnvAsDuration("CACHE_TTL", 24*time.Hour),
		},
		Kafka: KafkaConfig{
			Brokers: getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}, ","),
			Topic:   getEnv("KAFKA_TOPIC", "orders"),
			GroupID: getEnv("KAFKA_GROUP_ID", "order-service"),
		},
		HTTP: HTTPConfig{
			Addr: getEnv("HTTP_ADDR", ":8080"),
		},
		Consumer: ConsumerConfig{
			Timeout:    getEnvAsDuration("CONSUMER_TIMEOUT", 10*time.Second),
			MinBytes:   getEnvAsInt("CONSUMER_MIN_BYTES", 10*1024),      // 10KB
			MaxBytes:   getEnvAsInt("CONSUMER_MAX_BYTES", 10*1024*1024), // 10MB
			MaxWait:    getEnvAsDuration("CONSUMER_MAX_WAIT", 1*time.Second),
			RetryDelay: getEnvAsDuration("CONSUMER_RETRY_DELAY", 5*time.Second),
		},
		App: AppConfig{
			Env:      getEnv("APP_ENV", "development"),
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
		Validation: ValidationConfig{
	    Enabled:           getEnvAsBool("VALIDATION_ENABLED", true),
	    RequireOrderUID:   getEnvAsBool("VALIDATION_REQUIRE_ORDER_UID", true),
	    RequireTrackNumber: getEnvAsBool("VALIDATION_REQUIRE_TRACK_NUMBER", true),
	    RequireEntry:      getEnvAsBool("VALIDATION_REQUIRE_ENTRY", true),
	    RequireCustomerID: getEnvAsBool("VALIDATION_REQUIRE_CUSTOMER_ID", true),
	    RequireDelivery:   getEnvAsBool("VALIDATION_REQUIRE_DELIVERY", true),
	    RequirePayment:    getEnvAsBool("VALIDATION_REQUIRE_PAYMENT", true),
	    RequireItems:      getEnvAsBool("VALIDATION_REQUIRE_ITEMS", true),
	    ValidateEmail:     getEnvAsBool("VALIDATION_VALIDATE_EMAIL", true),
	    ValidatePhone:     getEnvAsBool("VALIDATION_VALIDATE_PHONE", true),
	    ValidateAmounts:   getEnvAsBool("VALIDATION_VALIDATE_AMOUNTS", true),
	    ValidateDates:     getEnvAsBool("VALIDATION_VALIDATE_DATES", true),
	    MaxItems:          getEnvAsInt("VALIDATION_MAX_ITEMS", 100),
	    MinAmount:         getEnvAsInt("VALIDATION_MIN_AMOUNT", 0),
	    MaxAmount:         getEnvAsInt("VALIDATION_MAX_AMOUNT", 100000000),
	    MaxDateFuture:     getEnvAsDuration("VALIDATION_MAX_DATE_FUTURE", 24*time.Hour),
	    MaxDatePast:       getEnvAsDuration("VALIDATION_MAX_DATE_PAST", 365*24*time.Hour),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database config: %w", err)
	}
	if err := c.Cache.Validate(); err != nil {
		return fmt.Errorf("cache config: %w", err)
	}
	if err := c.Kafka.Validate(); err != nil {
		return fmt.Errorf("kafka config: %w", err)
	}
	if err := c.Consumer.Validate(); err != nil {
		return fmt.Errorf("consumer config: %w", err)
	}
	return nil
}

func (c DatabaseConfig) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("POSTGRES_URL is required")
	}
	if c.MaxConns <= 0 {
		return fmt.Errorf("DB_MAX_CONNS must be positive")
	}
	if c.MinConns < 0 {
		return fmt.Errorf("DB_MIN_CONNS cannot be negative")
	}
	if c.MinConns > c.MaxConns {
		return fmt.Errorf("DB_MIN_CONNS cannot be greater than DB_MAX_CONNS")
	}
	return nil
}

func (c CacheConfig) Validate() error {
	if c.Addr == "" {
		return fmt.Errorf("REDIS_ADDR is required")
	}
	if c.TTL <= 0 {
		return fmt.Errorf("CACHE_TTL must be positive")
	}
	return nil
}

func (c KafkaConfig) Validate() error {
	if len(c.Brokers) == 0 {
		return fmt.Errorf("at least one KAFKA_BROKER is required")
	}
	if c.Topic == "" {
		return fmt.Errorf("KAFKA_TOPIC is required")
	}
	if c.GroupID == "" {
		return fmt.Errorf("KAFKA_GROUP_ID is required")
	}
	return nil
}

func (c ConsumerConfig) Validate() error {
	if c.Timeout <= 0 {
		return fmt.Errorf("CONSUMER_TIMEOUT must be positive")
	}
	if c.MinBytes <= 0 {
		return fmt.Errorf("CONSUMER_MIN_BYTES must be positive")
	}
	if c.MaxBytes <= c.MinBytes {
		return fmt.Errorf("CONSUMER_MAX_BYTES must be greater than CONSUMER_MIN_BYTES")
	}
	if c.MaxWait <= 0 {
		return fmt.Errorf("CONSUMER_MAX_WAIT must be positive")
	}
	return nil
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

func getEnvAsInt32(key string, defaultValue int32) int32 {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			return int32(intValue)
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}