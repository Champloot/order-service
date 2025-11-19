package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"order-service/internal/database"
	"order-service/internal/models"
)

func main() {
	ctx := context.Background()

	config := database.DatabaseConfig{
		URL:               "postgres://user:password@localhost:5432/orderservice?sslmode=disable",
		MaxConns:          5,
		MinConns:          2,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
	}

	repo, err := database.NewPostgresRepository(ctx, config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer repo.Close()

	orders := createTestOrders()
	
	for i := range orders {
		err := repo.SaveOrder(ctx, &orders[i])
		if err != nil {
			log.Printf("Failed to save order %s: %v", orders[i].OrderUID, err)
			continue
		}
		log.Printf("Order %s saved successfully", orders[i].OrderUID)
	}

	log.Println("Test data seeded successfully")
}

func createTestOrders() []models.Order {
	var orders []models.Order
	
	// creation of 10 test order
	for i := 1; i <= 10; i++ {
		order := models.Order{
			OrderUID:    fmt.Sprintf("test-order-%d", i),
			TrackNumber: fmt.Sprintf("WBILMTESTTRACK%d", i),
			Entry:       "WBIL",
			Delivery: models.Delivery{
				Name:    fmt.Sprintf("Test User %d", i),
				Phone:   "+9720000000",
				Zip:     "2639809",
				City:    "Kiryat Mozkin",
				Address: fmt.Sprintf("Test Address %d", i),
				Region:  "Kraiot",
				Email:   fmt.Sprintf("test%d@gmail.com", i),
			},
			Payment: models.Payment{
				Transaction:  fmt.Sprintf("b563feb7b2b84b6test%d", i),
				RequestID:    "",
				Currency:     "USD",
				Provider:     "wbpay",
				Amount:       1817 + i,
				PaymentDt:    1637907727,
				Bank:         "alpha",
				DeliveryCost: 1500,
				GoodsTotal:   317 + i,
				CustomFee:    0,
			},
			Items: []models.Item{
				{
					ChrtID:      9934930 + i,
					TrackNumber: fmt.Sprintf("WBILMTESTTRACK%d", i),
					Price:       453 + i,
					Rid:         fmt.Sprintf("ab4219087a764ae0btest%d", i),
					Name:        fmt.Sprintf("Test Product %d", i),
					Sale:        30,
					Size:        "0",
					TotalPrice:  317 + i,
					NmID:        2389212 + i,
					Brand:       "Vivienne Sabo",
					Status:      202,
				},
			},
			Locale:            "en",
			InternalSignature: "",
			CustomerID:        "test",
			DeliveryService:   "meest",
			Shardkey:          "9",
			SmID:              99,
			DateCreated:       time.Now(),
			OofShard:          "1",
		}
		orders = append(orders, order)
	}
	
	return orders
}
