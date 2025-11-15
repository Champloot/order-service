// +build integration

package cache_test

import (
	"context"
	"testing"
	"time"

	"order-service/internal/cache"
	"order-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_Integration(t *testing.T) {
	redisCache, err := cache.NewRedisCache("localhost:6379", "", 0, time.Hour)
	require.NoError(t, err, "REDIS NOT AVAILABLE - start containers with 'make docker-up'")
	defer redisCache.Close()

	ctx := context.Background()

	order := &models.Order{
		OrderUID:    "redis-test-" + time.Now().Format("20060102150405"),
		TrackNumber: "REDIS-TEST",
		Entry:       "TEST",
		DateCreated: time.Now(),
	}

	// Test Set
	err = redisCache.SetOrder(ctx, order)
	require.NoError(t, err, "Should set order in Redis")

	// Test Get
	retrieved, err := redisCache.GetOrder(ctx, order.OrderUID)
	require.NoError(t, err, "Should get order from Redis")
	assert.Equal(t, order.OrderUID, retrieved.OrderUID)

	// Test Delete
	err = redisCache.DeleteOrder(ctx, order.OrderUID)
	require.NoError(t, err, "Should delete order from Redis")

	// Verify deleted
	deleted, err := redisCache.GetOrder(ctx, order.OrderUID)
	require.NoError(t, err, "Should not error checking deleted order")
	assert.Nil(t, deleted, "Deleted order should be nil")
}