package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"order-service/internal/models"
	"order-service/internal/database"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader  *kafka.Reader
	db      *database.PostgresRepository
	timeout time.Duration
}

func NewConsumer(brokers []string, topic string, db *database.PostgresRepository) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		StartOffset: kafka.FirstOffset, // read from begin
		MinBytes:    10e3,              // 10KB
		MaxBytes:    10e6,              // 10MB
		MaxWait:     1 * time.Second,
	})

	return &Consumer{
		reader:  reader,
		db:      db,
		timeout: 10 * time.Second,
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
				time.Sleep(5 * time.Second) // Пауза перед повторной попыткой
				continue
			}

			log.Printf("Received message: %s", string(msg.Value))

			if err := c.processMessage(ctx, msg.Value); err != nil {
				log.Printf("Error processing message: %v", err)
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, data []byte) error {
	log.Printf("Processing raw message: %s", string(data))
	
	var order models.Order
	if err := json.Unmarshal(data, &order); err != nil {
		log.Printf("Failed to unmarshal order: %v", err)
		return fmt.Errorf("failed to unmarshal order: %v", err)
	}

	// Data validation
	if order.OrderUID == "" {
		log.Printf("Order UID is empty")
		return fmt.Errorf("order UID is required")
	}

	log.Printf("Processing order: %s", order.OrderUID)

	// save
	if err := c.db.SaveOrder(ctx, &order); err != nil {
		log.Printf("Failed to save order to database: %v", err)
		return fmt.Errorf("failed to save order to database: %v", err)
	}

	log.Printf("Successfully processed order %s", order.OrderUID)
	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
