package database_test

import (
	"context"
	"testing"
	"time"

	"order-service/internal/database"
	"order-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresRepository_Integration(t *testing.T) {
	ctx := context.Background()
	
	config := database.DatabaseConfig{
		URL:               "postgres://user:password@localhost:5432/orderservice?sslmode=disable",
		MaxConns:          5,
		MinConns:          2,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
	}
	
	// Пытаемся подключиться к БД
	repo, err := database.NewPostgresRepository(ctx, config)
	require.NoError(t, err, "DATABASE NOT AVAILABLE - start containers with 'make docker-up'")
	defer repo.Close()

	// Проверяем базовую функциональность
	order := &models.Order{
		OrderUID:    "integration-test-" + time.Now().Format("20060102150405"),
		TrackNumber: "INTEGRATION-TEST",
		Entry:       "TEST",
		Delivery: models.Delivery{
			Name:    "Integration Test",
			Phone:   "+1234567890",
			Zip:     "12345",
			City:    "Test City",
			Address: "Test Address",
			Region:  "Test Region",
			Email:   "test@example.com",
		},
		Payment: models.Payment{
			Transaction:  "integration-tx",
			Currency:     "USD",
			Provider:     "test",
			Amount:       1000,
			PaymentDt:    time.Now().Unix(),
			Bank:         "test-bank",
			DeliveryCost: 500,
			GoodsTotal:   500,
		},
		Items: []models.Item{
			{
				ChrtID:      111222,
				TrackNumber: "INTEGRATION-TEST",
				Price:       500,
				Rid:         "integration-rid",
				Name:        "Test Item",
				Sale:        0,
				Size:        "M",
				TotalPrice:  500,
				NmID:        333444,
				Brand:       "Test Brand",
				Status:      1,
			},
		},
		Locale:            "en",
		CustomerID:        "integration-customer",
		DeliveryService:   "test-service",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}

	// Test Save
	err = repo.SaveOrder(ctx, order)
	require.NoError(t, err, "Should save order to database")

	// Test Get
	retrieved, err := repo.GetOrder(ctx, order.OrderUID)
	require.NoError(t, err, "Should retrieve order from database")
	assert.Equal(t, order.OrderUID, retrieved.OrderUID)
	assert.Equal(t, order.TrackNumber, retrieved.TrackNumber)

	// Test GetAll
	orders, err := repo.GetAllOrders(ctx)
	require.NoError(t, err, "Should get all orders from database")
	assert.GreaterOrEqual(t, len(orders), 1, "Should have at least one order")

	// Test Delete
	err = repo.DeleteOrder(ctx, order.OrderUID)
	require.NoError(t, err, "Should delete order from database")

	// Verify deleted
	deleted, err := repo.GetOrder(ctx, order.OrderUID)
	assert.Error(t, err, "Should return error for deleted order")
	assert.Nil(t, deleted, "Deleted order should be nil")
}