package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"order-service/internal/models"
	"order-service/internal/ports"

	"github.com/segmentio/kafka-go"
)

var _ ports.OrderConsumer = (*Consumer)(nil)

type Consumer struct {
	reader		*kafka.Reader
	repository	ports.OrderRepository
	timeout		time.Duration
	retryDelay	time.Duration
}

func NewConsumer(
	brokers []string,
	topic string,
	repository ports.OrderRepository,
	timeout time.Duration,
	minBytes int,
	maxBytes int,
	maxWait time.Duration,
	retryDelay time.Duration,
) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		StartOffset: kafka.FirstOffset,
		MinBytes:    minBytes,
		MaxBytes:    maxBytes,
		MaxWait:     maxWait,
	})

	return &Consumer{
		reader:  	reader,
		repository:	repository,
		timeout: 	timeout,
		retryDelay:	retryDelay,
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
				time.Sleep(c.retryDelay) // Пауза перед повторной попыткой
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
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	log.Printf("Processing raw message: %s", string(data))
	
	var order models.Order
	if err := json.Unmarshal(data, &order); err != nil {
		log.Printf("Failed to unmarshal order: %v", err)
		return fmt.Errorf("Failed to unmarshal order: %w", err)
	}

	// Data validation
	if order.OrderUID == "" {
		log.Printf("Order UID is empty")
		return fmt.Errorf("Order UID is required")
	}

	log.Printf("Processing order: %s", order.OrderUID)

	// save
	err := c.repository.WithTransaction(ctx, func(tx ports.OrderTx) error {
		if err := tx.SaveOrder(ctx, &order); err != nil {
			return fmt.Errorf("Failed to save order to database: %w", err)
		}

		log.Printf("Successfully %s saved in transaction", order.OrderUID)
		return nil
	})

	if err != nil {
		log.Printf("Failed to sace order: %v", err)
		return fmt.Errorf("Failed to save order: %w", err)
	}

	log.Printf("Successfully processed order %s", order.OrderUID)
	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
