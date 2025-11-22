package cache

import (
	"context"
	"log"
	"encoding/json"
	"fmt"
	"time"
	"errors"

	"order-service/internal/models"
	"order-service/internal/ports"

	"github.com/go-redis/redis/v8"
)

var _ ports.OrderCache = (*RedisCache)(nil)

type RedisCache struct {
	client	*redis.Client
	ttl		time.Duration
}

func NewRedisCache(addr string, password string, db int, ttl time.Duration) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:		addr,
		Password:	password,
		DB:			db,
	})

	// Check connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Redis: %v", err)
	}

	return &RedisCache{
		client: client,
		ttl:    ttl,
	}, nil
}

func (c *RedisCache) SetOrder(ctx context.Context, order *models.Order) error {
	key := fmt.Sprintf("order:%s", order.OrderUID)
	jsonData, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("Failed to marshal order: %v", err)
	}

	err = c.client.Set(ctx, key, jsonData, c.ttl).Err()
	if err != nil {
		return fmt.Errorf("Failed to set order in cache: %v", err)
	}

	return nil
}

func (c *RedisCache) GetOrder(ctx context.Context, orderUID string) (*models.Order, error) {
	key := fmt.Sprintf("order:%s", orderUID)
	jsonData, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("Failed to get order from cache: %v", err)
	}

	var order models.Order
	err = json.Unmarshal(jsonData, &order)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal order: %v", err)
	}

	return &order, nil
}

func (c *RedisCache) DeleteOrder(ctx context.Context, orderUID string) error {
	key := fmt.Sprintf("order:%s", orderUID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("Failed to delete order from cache: %w", err)
	}
	return nil
}

func (c *RedisCache) PreloadOrders(ctx context.Context, orders []models.Order) error {
    var errs []error
    successCount := 0
    
    for _, order := range orders {
        if err := c.SetOrder(ctx, &order); err != nil {
            errs = append(errs, fmt.Errorf("order %s: %w", order.OrderUID, err))
            continue
        }
        successCount++
    }
    
    log.Printf("Preloaded %d/%d orders into cache", successCount, len(orders))
    
    if len(errs) > 0 {
        return fmt.Errorf("Failed to preload %d orders: %v", len(errs), errors.Join(errs...))
    }
    return nil
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}
