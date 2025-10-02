package main

import (
    "context"
    "encoding/json"
    "log"
    "time"

    "order-service/internal/models"

    "github.com/segmentio/kafka-go"
)

func main() {
    writer := &kafka.Writer{
        Addr:         kafka.TCP("localhost:9092"),
        Topic:        "orders",
        Balancer:     &kafka.LeastBytes{},
        BatchTimeout: 10 * time.Millisecond,
    }
    defer writer.Close()

    // test order
    order := models.Order{
        OrderUID:    "test-order-123",
        TrackNumber: "WBILMTESTTRACK",
        Entry:       "WBIL",
        Delivery: models.Delivery{
            Name:    "Test Testov",
            Phone:   "+9720000000",
            Zip:     "2639809",
            City:    "Kiryat Mozkin",
            Address: "Ploshad Mira 15",
            Region:  "Kraiot",
            Email:   "test@gmail.com",
        },
        Payment: models.Payment{
            Transaction:  "b563feb7b2b84b6test",
            RequestID:    "",
            Currency:     "USD",
            Provider:     "wbpay",
            Amount:       1817,
            PaymentDt:    1637907727,
            Bank:         "alpha",
            DeliveryCost: 1500,
            GoodsTotal:   317,
            CustomFee:    0,
        },
        Items: []models.Item{
            {
                ChrtID:      9934930,
                TrackNumber: "WBILMTESTTRACK",
                Price:       453,
                Rid:         "ab4219087a764ae0btest",
                Name:        "Mascaras",
                Sale:        30,
                Size:        "0",
                TotalPrice:  317,
                NmID:        2389212,
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

    orderJSON, err := json.Marshal(order)
    if err != nil {
        log.Fatalf("Failed to marshal order: %v", err)
    }

    err = writer.WriteMessages(context.Background(),
        kafka.Message{
            Key:   []byte(order.OrderUID),
            Value: orderJSON,
        },
    )
    if err != nil {
        log.Fatalf("Failed to write message: %v", err)
    }

    log.Println("Message sent successfully")
}
