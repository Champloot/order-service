package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"order-service/internal/models"
	"order-service/internal/database"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/snappy"
)

type Consumer struct {
	reader	*kafka.Reader
	db		*database.PostgresDB
	timeout	time.Duration
}

func NewConsumer(config *KafkaConfig, db *database.PostgresDB) *Consumer {
	dialer := &kafka.Dialer{
		Timeout:	10 * time.Second,
		DualStack:	true,
	}

	reader := kafka.NewReader(kafka.ReadConfig{
		Brokers:			config.Brokers,
		Topic:				config.Topic,
		GroupID:			config.GroupID,
		Dialer:				dialer,
		MinBytes:			config.MinBytes,
		MaxBytes:			config.MaxBytes,
		MaxWait:			1 * time.Second,
		CompressionCodec:	snappy.NewCompressionCodec(),
	})

	return &Consumer{
		reader:		reader,
		db:			db,
		timeout:	10 * time.Second,
	}
}

