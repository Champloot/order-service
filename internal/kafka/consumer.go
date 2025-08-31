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

func (c *Consumer) Start(ctx context.Context) {
	log.Println("Starting Kafka consumer...")

	for {
		select {
			case <-ctx.Done():
				return
			default:
				msg, err := c.reader.ReadMessage(ctx)
				if err != nil {
					log.Printf("Error reading message: %v", err)
					continue
				}

				log.Printf("Recived message: %s", string(msg.Value))

				if err := c.processMessage(ctx, msg.Value); err != nil {
					log.Printf("Error processing message: %v", err)
				}
		}
	}
}


func (c *Consumer) processMessage(ctx context.Context, data []byte) error {
	var order models.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return fmt.Errorf("Failed to unmarshal order: %v", err)
	}

	// data validation
	if order.OrderUID == "" {
		return fmt.Errorf("Order UID is required")
	}

	// save
	if err := c.db.SaveOrder(ctx, &order); err != nil {
		return fmt.Errorf("Failed to save order to database: %v", err)		
	}

	log.Printf("Successfully processed order %s", order.OrderUID)
	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}