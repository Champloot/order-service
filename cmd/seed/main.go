package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"order-service/internal/models"

	"github.com/segmentio/kafka-go"
)

func main() {
	// Connect Kafka
	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "orders",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

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

		// json serializtion
		orderJSON, err := json.Marshal(order)
		if err != nil {
			log.Fatalf("Failed to marshal order: %v", err)
		}

		// message sending
		err = writer.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(order.OrderUID),
				Value: orderJSON,
			},
		)
		if err != nil {
			log.Fatalf("Failed to write message: %v", err)
		}

		log.Printf("Message sent successfully: %s", order.OrderUID)
		time.Sleep(100 * time.Millisecond) // delay
	}

	log.Println("All test messages sent successfully")
}
